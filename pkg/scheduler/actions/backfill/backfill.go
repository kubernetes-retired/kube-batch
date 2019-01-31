/*
Copyright 2018 The Kubernetes Authors.

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

package backfill

import (
	"github.com/golang/glog"

	"github.com/kubernetes-sigs/kube-batch/pkg/scheduler/api"
	"github.com/kubernetes-sigs/kube-batch/pkg/scheduler/framework"
)

type backfillAction struct {
	ssn *framework.Session
}

func New() *backfillAction {
	return &backfillAction{}
}

func (alloc *backfillAction) Name() string {
	return "backfill"
}

func (alloc *backfillAction) Initialize() {}

func (alloc *backfillAction) Execute(ssn *framework.Session) {
	glog.V(3).Infof("Enter Backfill ...")
	defer glog.V(3).Infof("Leaving Backfill ...")

	// remove the top dog tasks that are not ready to run
	unReadyTopDogs := make(map[api.JobID]*api.JobInfo)
	for _, node := range ssn.Nodes {
		for _, task := range node.Tasks {
			if task.Status != api.Allocated && task.Status != api.AllocatedOverBackfill {
				continue
			}
			if _, ok := ssn.TopDogReadyJobs[task.Job]; !ok {
				// change task status back to Pending
				job := ssn.Jobs[task.Job]
				if err := job.UpdateTaskStatus(task, api.Pending); err != nil {
					glog.Errorf("Failed to update task <%v/%v> status to %v in Session <%v>: %v",
						task.Namespace, task.Name, api.Pending, ssn.UID, err)
				}

				unReadyTopDogs[task.Job] = ssn.Jobs[task.Job]

				// remove the task from the node
				node.RemoveTask(task)
				glog.Infof("removed task %s from node, idle is %v", task.Name, node.Idle.MilliCPU)
			}
		}
	}

	// pick a task to back fill
	// hack: using the job created most recently for testing
	// TODO: better way to pick backfill job
	var backfillJob *api.JobInfo
	backfillJob = nil
	for _, job := range ssn.Jobs {
		if backfillJob == nil {
			backfillJob = job
			continue
		}

		// skip ready jobs
		if ssn.JobReady(job) {
			glog.Infof("ignored ready job %s for backfill", job.Name)
			continue
		}

		if _, found := unReadyTopDogs[job.UID]; found {
			glog.Infof("ignored non-ready top dog job %s for backfill", job.Name)
			continue
		}

		allPending := true
		for _, task := range job.Tasks {
			if task.Status != api.Pending {
				allPending = false
				break
			}
		}
		if !allPending {
			glog.Infof("ignored non-pending job %s for backfill", job.Name)
			continue
		}

		if job.CreationTimestamp.After(backfillJob.CreationTimestamp.Time) {
			backfillJob = job
		}
	}

	if backfillJob == nil {
		glog.Info("no job to backfill")
		return
	}

	glog.Infof("picked job %v to backfill", backfillJob)

	// allocate(backfill) the task
	for _, task := range backfillJob.TaskStatusIndex[api.Pending] {
		task.IsBackfill = true
		for _, node := range ssn.Nodes {
			if err := ssn.PredicateFn(task, node); err != nil {
				glog.V(3).Infof("Predicates failed for task <%s/%s> on node <%s>: %v",
					task.Namespace, task.Name, node.Name, err)
				continue
			}

			if task.Resreq.LessEqual(node.Idle) {
				glog.V(3).Infof("Binding Task <%v/%v> to node <%v>", task.Namespace, task.Name, node.Name)
				if err := ssn.Allocate(task, node.Name, false); err != nil {
					glog.Errorf("Failed to bind Task %v on %v in Session %v", task.UID, node.Name, ssn.UID)
					continue
				}
			}

			// TODO: should stop further task allocation here if job is ready?
		}
	}

	// TODO (k82cn): When backfill, it's also need to balance between Queues.
	for _, job := range ssn.Jobs {
		for _, task := range job.TaskStatusIndex[api.Pending] {
			if task.InitResreq.IsEmpty() {
				// As task did not request resources, so it only need to meet predicates.
				// TODO (k82cn): need to prioritize nodes to avoid pod hole.
				for _, node := range ssn.Nodes {
					// TODO (k82cn): predicates did not consider pod number for now, there'll
					// be ping-pong case here.
					if err := ssn.PredicateFn(task, node); err != nil {
						glog.V(3).Infof("Predicates failed for task <%s/%s> on node <%s>: %v",
							task.Namespace, task.Name, node.Name, err)
						continue
					}

					glog.V(3).Infof("Binding Task <%v/%v> to node <%v>", task.Namespace, task.Name, node.Name)
					if err := ssn.Allocate(task, node.Name, false); err != nil {
						glog.Errorf("Failed to bind Task %v on %v in Session %v", task.UID, node.Name, ssn.UID)
						continue
					}
					break
				}
			} else {
			}
		}
	}
}

func (alloc *backfillAction) UnInitialize() {}
