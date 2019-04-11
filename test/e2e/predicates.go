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

package e2e

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeletapi "k8s.io/kubernetes/pkg/kubelet/apis"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
)

var _ = Describe("Predicates E2E Test", func() {
	It("NodeAffinity", func() {
		context := initTestContext()
		defer cleanupTestContext(context)

		slot := oneCPU
		nodeName, rep := computeNode(context, oneCPU)
		Expect(rep).NotTo(Equal(0))

		affinity := &v1.Affinity{
			NodeAffinity: &v1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
					NodeSelectorTerms: []v1.NodeSelectorTerm{
						{
							MatchFields: []v1.NodeSelectorRequirement{
								{
									Key:      schedulerapi.NodeFieldSelectorKeyNodeName,
									Operator: v1.NodeSelectorOpIn,
									Values:   []string{nodeName},
								},
							},
						},
					},
				},
			},
		}

		job := &jobSpec{
			name: "na-job",
			tasks: []taskSpec{
				{
					img:      "nginx",
					req:      slot,
					min:      1,
					rep:      1,
					affinity: affinity,
				},
			},
		}

		_, pg := createJob(context, job)
		err := waitPodGroupReady(context, pg)
		checkError(context, err)

		pods := getPodOfPodGroup(context, pg)
		for _, pod := range pods {
			Expect(pod.Spec.NodeName).To(Equal(nodeName))
		}
	})

	It("Hostport", func() {
		context := initTestContext()
		defer cleanupTestContext(context)

		nn := clusterNodeNumber(context)

		job := &jobSpec{
			name: "hp-job",
			tasks: []taskSpec{
				{
					img:      "nginx",
					min:      int32(nn),
					req:      oneCPU,
					rep:      int32(nn * 2),
					hostport: 28080,
				},
			},
		}

		_, pg := createJob(context, job)

		err := waitTasksReady(context, pg, nn)
		checkError(context, err)

		err = waitTasksPending(context, pg, nn)
		checkError(context, err)
	})

	It("Pod Affinity", func() {
		context := initTestContext()
		defer cleanupTestContext(context)

		slot := oneCPU
		_, rep := computeNode(context, oneCPU)
		Expect(rep).NotTo(Equal(0))

		labels := map[string]string{"foo": "bar"}

		affinity := &v1.Affinity{
			PodAffinity: &v1.PodAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: labels,
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		}

		job := &jobSpec{
			name: "pa-job",
			tasks: []taskSpec{
				{
					img:      "nginx",
					req:      slot,
					min:      rep,
					rep:      rep,
					affinity: affinity,
					labels:   labels,
				},
			},
		}

		_, pg := createJob(context, job)
		err := waitPodGroupReady(context, pg)
		checkError(context, err)

		pods := getPodOfPodGroup(context, pg)
		// All pods should be scheduled to the same node.
		nodeName := pods[0].Spec.NodeName
		for _, pod := range pods {
			Expect(pod.Spec.NodeName).To(Equal(nodeName))
		}
	})

	It("Taints/Tolerations", func() {
		context := initTestContext()
		defer cleanupTestContext(context)

		taints := []v1.Taint{
			{
				Key:    "test-taint-key",
				Value:  "test-taint-val",
				Effect: v1.TaintEffectNoSchedule,
			},
		}

		err := taintAllNodes(context, taints)
		checkError(context, err)

		job := &jobSpec{
			name: "tt-job",
			tasks: []taskSpec{
				{
					img: "nginx",
					req: oneCPU,
					min: 1,
					rep: 1,
				},
			},
		}

		_, pg := createJob(context, job)
		err = waitPodGroupPending(context, pg)
		checkError(context, err)

		err = removeTaintsFromAllNodes(context, taints)
		checkError(context, err)

		err = waitPodGroupReady(context, pg)
		checkError(context, err)
	})

	// Description: This test is to validate if the scheduler respects the max-pods limit of a node.
	// steps:
	// 1) Fetch a node from the list of available node.
	// 2) Get the nodes max-pods limit.
	// 3) Check nodes stability and get the pods running on that node.
	// 4) From the max-pods limit and the total pods runnnin from (2) and (3) we can calculate how many pods we can create.
	// 5) Create the remaining pods as a podgroup and deploy.
	// 6) Now create a single podgroup called pending and create one task with one replica [pod], this will remain in
	//    pending state since we have exhausted all the max pods limit.
	//
	It("validates MaxPods limit [Slow]", func() {
		context := initTestContext()
		defer func() {

			// need to make sure if the namespace is delete properly.
			// this takes time for this test case since we create lot may pods in podgroup
			cleanupTestContext(context)
		}()

		var totalPodCapacity int64
		var podsNeededForSaturation int32
		schedulableNodes := getAllWorkerNodes(context)
		nodeName := schedulableNodes[0].Name

		// get the max-pods capicity
		podCapacity, found := schedulableNodes[0].Status.Capacity[v1.ResourcePods]
		Expect(found).To(Equal(true))
		totalPodCapacity = podCapacity.Value()
		currentlyScheduledPods := getScheduledPodsOfNode(context, nodeName)
		podsNeededForSaturation = int32(totalPodCapacity) - int32(currentlyScheduledPods)
		if podsNeededForSaturation > 0 {
			// create a podgroup with nodeselector to the node and the number of replicas
			// in the task as podsNeededForSaturation

			affinityNode := &v1.Affinity{
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
			}

			maxpodJobOne := &jobSpec{
				name: "max-pods",
				tasks: []taskSpec{
					{
						img:      "nginx",
						min:      podsNeededForSaturation,
						rep:      podsNeededForSaturation,
						affinity: affinityNode,
					},
				},
			}

			//Schedule Job in the selected Node
			By(fmt.Sprintf("Creating the PodGroup with %v task replicas on node %v", podsNeededForSaturation, nodeName))
			_, maxpodspg := createJob(context, maxpodJobOne)
			By("Waiting for the PodGroup to have running status")
			err := waitTimeoutPodGroupReady(context, maxpodspg, fiveMinutes)
			checkError(context, err)

		}
		// create the new podgroup with on task replica on the same node as created before.
		// this podgroup should remain pending
		affinityNode := &v1.Affinity{
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
		}
		maxpodJobTwo := &jobSpec{
			name: "unscheduled-pod",
			tasks: []taskSpec{
				{
					img:      "nginx",
					min:      1,
					rep:      1,
					affinity: affinityNode,
				},
			},
		}

		//Schedule Job in the selected Node
		By("Creating the PodGroup with one task replicas")
		_, pendingpg := createJob(context, maxpodJobTwo)
		By("Waiting for the PodGroup status to remain pending")
		err := waitTimeoutPodGroupReady(context, pendingpg, oneMinute)
		if err != wait.ErrWaitTimeout {
			checkError(context, err)
		}

	})
})
