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

package kube_batch

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("Job E2E Test", func() {
	It("Try to fit unassigned task with different resource requests in one loop", func() {
		context := initTestContext()
		defer cleanupTestContext(context)

		slot := oneCPU
		rep := clusterSize(context, slot)
		minMemberOverride := int32(1)

		replicaset := createReplicaSet(context, "rs-1", rep-1, "nginx", slot)
		err := waitReplicaSetReady(context, replicaset.Name)
		checkError(context, err)

		_, pg := createJob(context, &jobSpec{
			name: "multi-task-diff-resource-job",
			tasks: []taskSpec{
				{
					img: "nginx",
					pri: masterPriority,
					min: 1,
					rep: 1,
					req: twoCPU,
				},
				{
					img: "nginx",
					pri: workerPriority,
					min: 1,
					rep: 1,
					req: halfCPU,
				},
			},
			minMember: &minMemberOverride,
		})

		err = waitPodGroupPending(context, pg)
		checkError(context, err)

		// task_1 has been scheduled
		err = waitTasksReady(context, pg, int(minMemberOverride))
		checkError(context, err)
	})
})
