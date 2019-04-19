/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package reserve

import (
	"fmt"
	"time"

	"github.com/kubernetes-sigs/kube-batch/pkg/apis/scheduling/v1alpha1"

	"github.com/golang/glog"
	"github.com/kubernetes-sigs/kube-batch/pkg/scheduler/api"
	"github.com/kubernetes-sigs/kube-batch/pkg/scheduler/framework"
	"github.com/kubernetes-sigs/kube-batch/pkg/scheduler/util"
)

type reserveAction struct {
	ssn       *framework.Session
	arguments framework.Arguments
}

// New inits an action
func New(args framework.Arguments) framework.Action {
	return &reserveAction{
		arguments: args,
	}
}

func (alloc *reserveAction) Name() string {
	return "reserve"
}

func (alloc *reserveAction) Initialize() {}

func (alloc *reserveAction) Execute(ssn *framework.Session) {
	glog.V(3).Infof("Enter Reserve ...")
	defer glog.V(3).Infof("Leaving Reserve ...")

	enabled := false
	alloc.arguments.GetBool(&enabled, "enabled")
	if !enabled {
		glog.V(3).Infof("Reserve disabled")
		return
	}

	// get the starvation threshold time
	var threshold time.Duration
	err := alloc.getStarvationThreshold(&threshold)
	if err != nil {
		return
	}

	// collect starving jobs
	queues := util.NewPriorityQueue(ssn.QueueOrderFn)
	jobsMap := map[api.QueueID]*util.PriorityQueue{}

	for _, job := range ssn.Jobs {
		if queue, found := ssn.Queues[job.Queue]; found {
			queues.Push(queue)
		} else {
			glog.V(4).Infof("Skip adding job <%s/%s> because its queue %s is not found",
				job.Namespace, job.Name, job.Queue)
			continue
		}
		if !isJobStarving(job, threshold) {
			glog.V(4).Infof("Skip adding job <%s/%s> because it is not starving",
				job.Namespace, job.Name, job.Queue)
			continue
		}

		if _, found := jobsMap[job.Queue]; !found {
			jobsMap[job.Queue] = util.NewPriorityQueue(ssn.JobOrderFn)
		}

		glog.V(4).Infof("Added Job <%s/%s> into Queue <%s>", job.Namespace, job.Name, job.Queue)
		jobsMap[job.Queue].Push(job)
	}

	pendingTasks := map[api.JobID]*util.PriorityQueue{}
	allNodes := util.GetNodeList(ssn.Nodes)
	predicateFn := func(task *api.TaskInfo, node *api.NodeInfo) error {
		// a node is elligible for starving job if its allocatable >= task.InitResreq
		if !task.InitResreq.LessEqual(node.Allocatable) {
			return fmt.Errorf("task <%s/%s> ResourceFit failed on node <%s>",
				task.Namespace, task.Name, node.Name)
		}
		return ssn.PredicateFn(task, node)
	}

	// (over-)allocate resource to starving jobs
	for {
		if queues.Empty() {
			break
		}

		queue := queues.Pop().(*api.QueueInfo)
		if ssn.Overused(queue) {
			glog.V(3).Infof("Queue <%s> is overused, ignore it.", queue.Name)
			continue
		}

		jobs, found := jobsMap[queue.UID]
		if !found || jobs.Empty() {
			glog.V(4).Infof("Cannot find jobs for queue %s.", queue.Name)
			continue
		}

		// collect all pending tasks of the starving job
		job := jobs.Pop().(*api.JobInfo)
		if _, found := pendingTasks[job.UID]; !found {
			tasks := util.NewPriorityQueue(ssn.TaskOrderFn)
			for _, task := range job.TaskStatusIndex[api.Pending] {
				// Skip BestEffort task in 'allocate' action.
				if task.Resreq.IsEmpty() {
					glog.V(4).Infof("Task <%v/%v> is BestEffort task, skip it.",
						task.Namespace, task.Name)
					continue
				}

				tasks.Push(task)
			}
			pendingTasks[job.UID] = tasks
		}
		tasks := pendingTasks[job.UID]

		glog.V(3).Infof("Trying to reserve resource for %d tasks of starving Job <%v/%v>",
			tasks.Len(), job.Namespace, job.Name)

		for !tasks.Empty() {
			task := tasks.Pop().(*api.TaskInfo)

			glog.V(3).Infof("There are <%d> nodes for Job <%v/%v>",
				len(ssn.Nodes), job.Namespace, job.Name)

			if len(job.NodesFitDelta) > 0 {
				job.NodesFitDelta = make(api.NodeResourceMap)
			}

			// Going through all the nodes.
			// When allocating for starving, node needs
			// to be offered in consistent order
			// If no node is available, task will not be allocated
			// in any case
			nodes := util.PredicateNodes(task, allNodes, predicateFn)
			if len(nodes) == 0 {
				break
			}

			taskAllocated := false
			for _, node := range nodes {
				// Allocate idle resource to the task.
				if task.InitResreq.LessEqual(node.Idle) || toOverAllocate(ssn, node, threshold, task) {
					if err := ssn.Allocate(task, node); err != nil {
						glog.Errorf("Failed to bind task %v on %v in Session %v, err: %v",
							task.UID, node.Name, ssn.UID, err)
					}
					taskAllocated = true
					glog.V(3).Infof("Bond task <%v/%v> to node <%v>",
						task.Namespace, task.Name, node.Name)
					break
				} else {
					job.NodesFitDelta[node.Name] = node.Idle.Clone()
					job.NodesFitDelta[node.Name].FitDelta(task.InitResreq)
				}
			}

			// stop trying for this job if no (more) nodes available for this task
			if !taskAllocated {
				break
			}

			if ssn.JobReady(job) {
				jobs.Push(job)
				break
			}
		}

		// Added Queue back until no job in Queue.
		queues.Push(queue)
	}
}

// ToOverAllocate check if it is okay to allocate resource on given node ignoring
// running tasks
func toOverAllocate(ssn *framework.Session, node *api.NodeInfo,
	StarvationThreshold time.Duration, task *api.TaskInfo) bool {
	job, found := ssn.Jobs[task.Job]
	if !found {
		return false
	}
	isJobStarving := isJobStarving(job, StarvationThreshold)
	if !isJobStarving {
		return false
	}

	// allow over-allocate for starving job
	netResource := node.Allocatable.Clone()
	for _, nodeTask := range node.Tasks {
		if nodeTask.Status == api.Borrowing ||
			nodeTask.Status == api.Allocated {
			netResource.Sub(nodeTask.InitResreq)
		}
	}

	// return true means when allocating for starving job, over-allocate resources ignoring
	// resources being used on the node to prevent other non-startving job from taking the resources
	return !task.InitResreq.LessEqual(node.Idle) &&
		task.InitResreq.LessEqual(netResource)
}

func isJobStarving(jobInfo *api.JobInfo, starvationThreshold time.Duration) bool {
	readiness := jobInfo.GetReadiness()
	return readiness != api.Ready &&
		readiness != api.ConditionallyReady &&
		time.Since(jobInfo.CreationTimestamp.Time) >= starvationThreshold
}

func (alloc *reserveAction) getStarvationThreshold(threshold *time.Duration) error {
	zeroDuration, _ := time.ParseDuration("0h")
	*threshold = zeroDuration
	alloc.arguments.GetTimeDuration(threshold, "threshold")
	if *threshold == zeroDuration {
		var err error
		*threshold, err = time.ParseDuration(v1alpha1.DefaultStarvingThreshold)
		if err != nil {
			glog.Error("no valid default starvation setting: %v", v1alpha1.DefaultStarvingThreshold)
			return err
		}
		glog.Infof("no starvation threshold time in config. using default value %v", threshold)
	}

	return nil
}

func (alloc *reserveAction) UnInitialize() {}
