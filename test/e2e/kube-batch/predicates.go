/*
Copyright 2019 The Kubernetes Authors.

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
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	kubev1 "github.com/kubernetes-sigs/kube-batch/pkg/apis/scheduling/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeletapi "k8s.io/kubernetes/pkg/kubelet/apis"
)

var _ = Describe("Predicates E2E Test [kube-batch]", func() {

	// This test verifies we don't allow scheduling of jobs in a way that sum of
	// limits of tasks is greater than machines capacity.

	It("should validates resource limits of tasks that are allowed to run ", func() {
		context := initTestContext()
		defer cleanupTestContext(context)

		// The max cpu available across the nodes
		nodeMaxAllocatableCpu := int64(0)
		// The max mem available across the nodes
		nodeMaxAllocatableMem := int64(0)
		// Map of available cpu in a node
		nodeToAllocatableMapCpu := make(map[string]int64)
		// Map of available mem in a node
		nodeToAllocatableMemMap := make(map[string]int64)
		nodeList, err := context.kubeclient.CoreV1().Nodes().List(metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		for _, node := range nodeList.Items {
			nodeReady := false
			for _, condition := range node.Status.Conditions {
				if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
					nodeReady = true
					break
				}
			}

			if IsMasterNode(&node) {
				continue
			}
			// Skip the unready node.
			if !nodeReady {
				continue
			}
			// Find allocatable amount of CPU.
			allocatable, found := node.Status.Allocatable[v1.ResourceCPU]
			Expect(found).To(Equal(true))
			nodeToAllocatableMapCpu[node.Name] = allocatable.MilliValue()
			if nodeMaxAllocatableCpu < allocatable.MilliValue() {
				nodeMaxAllocatableCpu = allocatable.MilliValue()
			}

			// Find allocatable amount of Memory.
			allocatableMem, found := node.Status.Allocatable[v1.ResourceMemory]
			Expect(found).To(Equal(true))
			nodeToAllocatableMemMap[node.Name] = allocatableMem.Value()
			if nodeMaxAllocatableMem < allocatableMem.Value() {
				nodeMaxAllocatableMem = allocatableMem.Value()
			}
		}

		pods, err := context.kubeclient.CoreV1().Pods(metav1.NamespaceAll).List(metav1.ListOptions{})
		checkError(context, err)
		for _, pod := range pods.Items {
			_, found := nodeToAllocatableMapCpu[pod.Spec.NodeName]
			_, foundMem := nodeToAllocatableMemMap[pod.Spec.NodeName]
			if found && foundMem && pod.Status.Phase != v1.PodSucceeded && pod.Status.Phase != v1.PodFailed {
				nodeToAllocatableMapCpu[pod.Spec.NodeName] -= getRequestedCPU(pod)
				nodeToAllocatableMemMap[pod.Spec.NodeName] -= getRequestedMemory(pod)
			}
		}

		By("Starting Jobs to consume most of the cluster CPU.")
		// Create a job per node ( i.e with a new podgroup per job).
		// These jobs will exhaust 70% of the CPU on that node.
		// Using node-affinity assign the job per node.

		cpuPodGrpList := []*kubev1.PodGroup{}
		for nodeName, cpu := range nodeToAllocatableMapCpu {
			requestedCPU := cpu * 7 / 10
			cpuJob := &jobSpec{
				name: fmt.Sprintf("cpu-filler-job-%s", nodeName),
				pri:  masterPriority,
				tasks: []taskSpec{
					{
						img: "nginx",
						req: v1.ResourceList{
							v1.ResourceCPU: *resource.NewMilliQuantity(requestedCPU, resource.DecimalSI),
						},
						min: 1,
						rep: 1,
						affinity: &v1.Affinity{
							NodeAffinity: &v1.NodeAffinity{
								RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
									NodeSelectorTerms: []v1.NodeSelectorTerm{
										{
											MatchExpressions: []v1.NodeSelectorRequirement{
												{
													Key:      kubeletapi.LabelHostname,
													Operator: v1.NodeSelectorOpIn,
													Values:   []string{nodeName},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}
			_, cpuPodGroup := createJob(context, cpuJob)
			cpuPodGrpList = append(cpuPodGrpList, cpuPodGroup)
		}
		// Wait for CPU filler jobs to be scheduled.
		for _, cpuPodGroup := range cpuPodGrpList {
			err = waitTimeoutPodGroupReady(context, cpuPodGroup, tenMinute)
			checkError(context, err)
		}

		By("Starting Jobs to consume most of the cluster Memory.")
		// Create a job per node ( i.e with a new podgroup per job).
		// These jobs will exhaust 70% of the MEM on that node.
		// Using node-affinity assign the job per node.

		memPodGrpList := []*kubev1.PodGroup{}
		for nodeName, memory := range nodeToAllocatableMemMap {
			requestedMem := memory * 7 / 10
			memJob := &jobSpec{
				name: fmt.Sprintf("mem-filler-job-%s", nodeName),
				pri:  masterPriority,
				tasks: []taskSpec{
					{
						img: "nginx",
						req: v1.ResourceList{
							v1.ResourceMemory: *resource.NewMilliQuantity(requestedMem, resource.DecimalSI),
						},
						min: 1,
						rep: 1,
						affinity: &v1.Affinity{
							NodeAffinity: &v1.NodeAffinity{
								RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
									NodeSelectorTerms: []v1.NodeSelectorTerm{
										{
											MatchExpressions: []v1.NodeSelectorRequirement{
												{
													Key:      kubeletapi.LabelHostname,
													Operator: v1.NodeSelectorOpIn,
													Values:   []string{nodeName},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}
			_, memPodGroup := createJob(context, memJob)
			memPodGrpList = append(memPodGrpList, memPodGroup)

		}
		// Wait for mem filler Jobs to be scheduled.
		for _, memPod := range memPodGrpList {
			err = waitTimeoutPodGroupReady(context, memPod, tenMinute)
			checkError(context, err)
		}

		By("Creating another Job that requires unavailable amount of CPU.")
		// Create another Job that requires 50% of the largest node CPU resources.
		// This Job should remain pending as at least 70% of CPU of other nodes in
		// the cluster are already consumed.

		addCpuJob := &jobSpec{
			name: "additional-job-cpu",
			pri:  masterPriority,
			tasks: []taskSpec{
				{
					img: "nginx",
					req: v1.ResourceList{
						v1.ResourceCPU: *resource.NewMilliQuantity(((nodeMaxAllocatableCpu * 5) / 10), resource.DecimalSI),
					},
					min: 1,
					rep: 1,
				},
			},
		}
		_, addCpuPodGroup := createJob(context, addCpuJob)
		// this additional-job-cpu should be pending
		// check if the pod group is ready,
		// this check should result in a timeout, then the test case is successful.

		err = waitTimeoutPodGroupReady(context, addCpuPodGroup, oneMinute)
		if err != wait.ErrWaitTimeout {
			checkError(context, err)
		}

		By("Creating another Job that requires unavailable amount of Memory.")
		// Create another Job that requires 50% of the largest node Memory resources.
		// This Job should remain pending as at least 70% of Memory of other nodes in
		// the cluster are already consumed.
		addMemJob := &jobSpec{
			name: "additional-job-mem",
			pri:  masterPriority,
			tasks: []taskSpec{
				{
					img: "nginx",
					req: v1.ResourceList{
						v1.ResourceMemory: *resource.NewMilliQuantity(((nodeMaxAllocatableCpu * 5) / 10), resource.DecimalSI),
					},
					min: 1,
					rep: 1,
				},
			},
		}
		_, addMemPodGroup := createJob(context, addMemJob)
		// This additional-job-mem should be pending
		// check if the pod group is ready,
		// this check should result in a timeout, then the test case is successful.

		err = waitTimeoutPodGroupReady(context, addMemPodGroup, oneMinute)
		if err != wait.ErrWaitTimeout {
			checkError(context, err)
		}
	})

})
