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
					if err := ssn.Allocate(task, node); err != nil {
						glog.Errorf("Failed to bind Task %v on %v in Session %v", task.UID, node.Name, ssn.UID)
						continue
					}
					break
				}
			} else {
				// TODO (k82cn): backfill for other case.
			}
		}
	}

	if ssn.EnableBackfill {
		// Collect back fill candidates
		backFillCandidates := make([]*api.JobInfo, 0, len(ssn.Jobs))
		for _, job := range ssn.Jobs {
			if !ssn.BackFillEligible(job) {
				continue
			}
			backFillCandidates = append(backFillCandidates, job)
		}

		// Release resources allocated to unready top dog jobs so that
		// we can back fill more jobs in the next step.
		resourceReleased := false
		for _, job := range ssn.Jobs {
			if ssn.JobReady(job) || job.GetReadiness() == api.OverResourceReady || job.Starving(ssn.StarvationThreshold) {
				glog.V(3).Infof("Disable backfill on job <%v/%v>", job.Namespace, job.Name)
			} else {
				releaseReservedResources(ssn, job)
				resourceReleased = true
			}
		}

		if resourceReleased {
			for _, job := range backFillCandidates {
				backFill(ssn, job)
			}
		}
	}
}

// Releases resources allocated to the given job back to the cluster.
func releaseReservedResources(ssn *framework.Session, job *api.JobInfo) {
	glog.V(3).Infof("Releasing resources allocated to job <%v/%v>", job.Namespace, job.Name)

	for _, task := range job.Tasks {
		if task.Status == api.Allocated || task.Status == api.AllocatedOverBackfill {
			if err := job.UpdateTaskStatus(task, api.Pending); err != nil {
				glog.Errorf("Failed to update task <%v/%v> status to %v in Session <%v>: %v",
					task.Namespace, task.Name, api.Pending, ssn.UID, err)
			}

			node := ssn.Nodes[task.NodeName]
			if err := node.RemoveTask(task); err != nil {
				glog.Errorf("Failed to remove task %v from node %v: %s", task.Name, node.Name, err)
				continue
			}

			glog.V(4).Infof("Removed task %s from node %s. Idle: %+v; Used: %v; Releasing: %v.", task.Name, node.Name, node.Idle, node.Used, node.Releasing)
		}
	}
}

func backFill(ssn *framework.Session, job *api.JobInfo) {
	glog.V(3).Infof("Backfill job <%v/%v>", job.Namespace, job.Name)

	for _, task := range job.TaskStatusIndex[api.Pending] {
		allocateFailed := false
		for _, node := range ssn.Nodes {
			if err := ssn.PredicateFn(task, node); err != nil {
				glog.V(3).Infof("Predicates failed for task <%s/%s> on node <%s>: %v",
					task.Namespace, task.Name, node.Name, err)
				continue
			}

			if task.InitResreq.LessEqual(node.Idle) {
				task.IsBackfill = true
				glog.V(3).Infof("Binding backfill task <%v/%v> to node <%v>", task.Namespace, task.Name, node.Name)
				if err := ssn.Allocate(task, node); err != nil {
					glog.Errorf("Failed to bind backfill task %v on %v in Session %v: %s", task.UID, node.Name, ssn.UID, err)
					allocateFailed = true
				}
				break
			}
		}
		if allocateFailed {
			glog.V(3).Infof("Aborted backilling job %s", job.Name)
			break
		}
	}

	if !ssn.JobReady(job) {
		glog.V(3).Infof("Job <%v/%v> is not ready. Release its resources.", job.Namespace, job.Name)
		releaseReservedResources(ssn, job)
	}
}

func (alloc *backfillAction) UnInitialize() {}
