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

package e2e

import (
	"fmt"
	"math/rand"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeletapi "k8s.io/kubernetes/pkg/kubelet/apis"
)

var _ = Describe("NodeOrder E2E Test", func() {
	It("Node Affinity Test", func() {
		context := initTestContext()
		defer cleanupTestContext(context)

		nodeNames := getAllWorkerNodes(context)
		var preferredSchedulingTermSlice []v1.PreferredSchedulingTerm
		nodeSelectorRequirement := v1.NodeSelectorRequirement{Key: kubeletapi.LabelHostname, Operator: v1.NodeSelectorOpIn, Values: []string{nodeNames[0]}}
		nodeSelectorTerm := v1.NodeSelectorTerm{MatchExpressions: []v1.NodeSelectorRequirement{nodeSelectorRequirement}}
		schedulingTerm := v1.PreferredSchedulingTerm{Weight: 100, Preference: nodeSelectorTerm}
		preferredSchedulingTermSlice = append(preferredSchedulingTermSlice, schedulingTerm)

		slot := oneCPU
		_, rep := computeNode(context, oneCPU)
		Expect(rep).NotTo(Equal(0))

		affinity := &v1.Affinity{
			NodeAffinity: &v1.NodeAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: preferredSchedulingTermSlice,
			},
		}

		job := &jobSpec{
			name: "pa-job",
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
		//All pods should be scheduled in particular node
		for _, pod := range pods {
			Expect(pod.Spec.NodeName).To(Equal(nodeNames[0]))
		}
	})

	It("Pod Affinity Test", func() {
		context := initTestContext()
		defer cleanupTestContext(context)

		var preferredSchedulingTermSlice []v1.WeightedPodAffinityTerm
		labelSelectorRequirement := metav1.LabelSelectorRequirement{Key: "test", Operator: metav1.LabelSelectorOpIn, Values: []string{"e2e"}}
		labelSelector := &metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{labelSelectorRequirement}}
		podAffinityTerm := v1.PodAffinityTerm{LabelSelector: labelSelector, TopologyKey: "kubernetes.io/hostname"}
		weightedPodAffinityTerm := v1.WeightedPodAffinityTerm{Weight: 100, PodAffinityTerm: podAffinityTerm}
		preferredSchedulingTermSlice = append(preferredSchedulingTermSlice, weightedPodAffinityTerm)

		labels := make(map[string]string)
		labels["test"] = "e2e"

		job1 := &jobSpec{
			name: "pa-job1",
			tasks: []taskSpec{
				{
					img:    "nginx",
					req:    halfCPU,
					min:    1,
					rep:    1,
					labels: labels,
				},
			},
		}

		_, pg1 := createJob(context, job1)
		err := waitPodGroupReady(context, pg1)
		checkError(context, err)

		pods := getPodOfPodGroup(context, pg1)
		nodeName := pods[0].Spec.NodeName

		affinity := &v1.Affinity{
			PodAffinity: &v1.PodAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: preferredSchedulingTermSlice,
			},
		}

		job2 := &jobSpec{
			name: "pa-job2",
			tasks: []taskSpec{
				{
					img:      "nginx",
					req:      halfCPU,
					min:      1,
					rep:      1,
					affinity: affinity,
				},
			},
		}

		_, pg2 := createJob(context, job2)
		err = waitPodGroupReady(context, pg2)
		checkError(context, err)

		podsWithAffinity := getPodOfPodGroup(context, pg2)
		// All Pods Should be Scheduled in same node
		nodeNameWithAffinity := podsWithAffinity[0].Spec.NodeName
		Expect(nodeNameWithAffinity).To(Equal(nodeName))

	})

	It("Least Requested Resource Test", func() {
		context := initTestContext()
		defer cleanupTestContext(context)

		nodeNames := getAllWorkerNodes(context)
		affinityNodeOne := &v1.Affinity{
			NodeAffinity: &v1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
					NodeSelectorTerms: []v1.NodeSelectorTerm{
						{
							MatchExpressions: []v1.NodeSelectorRequirement{
								{
									Key:      kubeletapi.LabelHostname,
									Operator: v1.NodeSelectorOpIn,
									Values:   []string{nodeNames[0]},
								},
							},
						},
					},
				},
			},
		}

		job1 := &jobSpec{
			name: "pa-job",
			tasks: []taskSpec{
				{
					img:      "nginx",
					req:      halfCPU,
					min:      3,
					rep:      3,
					affinity: affinityNodeOne,
				},
			},
		}

		//Schedule Job in first Node
		_, pg1 := createJob(context, job1)
		err := waitPodGroupReady(context, pg1)
		checkError(context, err)

		affinityNodeTwo := &v1.Affinity{
			NodeAffinity: &v1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
					NodeSelectorTerms: []v1.NodeSelectorTerm{
						{
							MatchExpressions: []v1.NodeSelectorRequirement{
								{
									Key:      kubeletapi.LabelHostname,
									Operator: v1.NodeSelectorOpIn,
									Values:   []string{nodeNames[1]},
								},
							},
						},
					},
				},
			},
		}

		job2 := &jobSpec{
			name: "pa-job1",
			tasks: []taskSpec{
				{
					img:      "nginx",
					req:      halfCPU,
					min:      3,
					rep:      3,
					affinity: affinityNodeTwo,
				},
			},
		}

		//Schedule Job in Second Node
		_, pg2 := createJob(context, job2)
		err = waitPodGroupReady(context, pg2)
		checkError(context, err)

		testJob := &jobSpec{
			name: "pa-test-job",
			tasks: []taskSpec{
				{
					img: "nginx",
					req: oneCPU,
					min: 1,
					rep: 1,
				},
			},
		}

		//This job should be scheduled in third node
		_, pg3 := createJob(context, testJob)
		err = waitPodGroupReady(context, pg3)
		checkError(context, err)

		pods := getPodOfPodGroup(context, pg3)
		for _, pod := range pods {
			Expect(pod.Spec.NodeName).NotTo(Equal(nodeNames[0]))
			Expect(pod.Spec.NodeName).NotTo(Equal(nodeNames[1]))
		}
	})

	// Description:
	// Test Nodes does not have any label, hence it should be impossible to schedule Pod with
	// nonempty Selector set.
	//
	It("validates that NodeSelector is respected if not matching ", func() {
		context := initTestContext()
		defer cleanupTestContext(context)
		By("Trying to schedule a Job with nonempty NodeSelector.")
		testJob := &jobSpec{
			name: "restricted-nodeselector-job",
			tasks: []taskSpec{
				{
					img: "nginx",
					req: oneCPU,
					min: 3,
					rep: 3,
					nodeselector: map[string]string{
						"label": "nonempty",
					},
				},
			},
		}

		// Create a Pod-Group Job
		_, nodeSelectorpg := createJob(context, testJob)
		// This PodGroup cannot be run and all the three task will be in pending state
		// since this test doesn't satisfy the nodeselector
		err := waitPodGroupPending(context, nodeSelectorpg)
		checkError(context, err)
	})

	// Description: Ensure that scheduler respects the NodeSelector field
	// of PodSpec during scheduling (when it matches).

	It("validates that NodeSelector is respected if matching ", func() {
		context := initTestContext()
		defer cleanupTestContext(context)
		// Randomly pick a node
		nodeName := getAllWorkerNodes(context)

		By("Trying to apply a random label on the found node.")
		labelKey := fmt.Sprintf("kubernetes.io/e2e-%d", rand.Uint32())
		labelValue := "42"

		err := addLabelsToNode(context, nodeName[0], map[string]string{labelKey: labelValue})
		checkError(context, err)
		expectNodeHasLabel(context, nodeName[0], labelKey, labelValue)

		By("Trying to create a Job with labels.")
		testJob := &jobSpec{
			name: "nodeselector-job",
			tasks: []taskSpec{
				{
					img: "nginx",
					req: v1.ResourceList{"cpu": resource.MustParse("1m")},
					min: 3,
					rep: 3,
					nodeselector: map[string]string{
						labelKey: labelValue,
					},
				},
			},
		}

		// Create the Pod-Group Job
		_, nodeSelectorpg := createJob(context, testJob)
		// check that job got scheduled.
		By("validate if Job match the node label")
		err = waitPodGroupReady(context, nodeSelectorpg)
		checkError(context, err)

		// check if the pods in the pod group are running in the correct node
		// fetch the pods from the job and check if it has right node name in spec.node.name

		By("Validate if the Job is running in the node having the label")
		podList := getPodOfPodGroup(context, nodeSelectorpg)
		for _, pod := range podList {
			Expect(pod.Spec.NodeName).To(Equal(nodeName[0]))
		}

		By("Remove the node label")
		removeLabelOffNode(context, nodeName[0], []string{labelKey})
		err = verifyLabelsRemoved(context, nodeName[0], []string{labelKey})
		checkError(context, err)
	})

	It("validates scheduler respect's a pod with requiredDuringSchedulingIgnoredDuringExecution constraint [Hard NodeAffinity]", func() {
		context := initTestContext()
		defer cleanupTestContext(context)

		nodeAffinityHard := &v1.Affinity{
			NodeAffinity: &v1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
					NodeSelectorTerms: []v1.NodeSelectorTerm{
						{
							MatchExpressions: []v1.NodeSelectorRequirement{
								{
									Key:      "gpu",
									Operator: v1.NodeSelectorOpIn,
									Values: []string{
										"nvidia",
										"zotac-nvidia",
									},
								},
							},
						},
					},
				},
			},
		}

		By("Trying to get a schedulable node")
		schedulableNodes := getAllWorkerNodes(context)
		nodeOne := schedulableNodes[0]

		By("Trying to apply a label on the found node.")
		k := "gpu"
		v := "nvidia"
		addLabelsToNode(context, nodeOne, map[string]string{k: v})
		expectNodeHasLabel(context, nodeOne, k, v)

		By("Trying to launch a Job, with requiredDuringSchedulingIgnoredDuringExecution constraint on its tasks")
		nodeAffinityHardjs := &jobSpec{
			name: "node-affinity-hard",
			tasks: []taskSpec{
				{
					img:      "nginx",
					min:      1,
					rep:      1,
					affinity: nodeAffinityHard,
				},
			},
		}
		_, nodeAffinityHardpg := createJob(context, nodeAffinityHardjs)
		err := waitPodGroupReady(context, nodeAffinityHardpg)
		checkError(context, err)

		By("Validate if the Job is running in the node having the label")
		podList := getPodOfPodGroup(context, nodeAffinityHardpg)
		for _, pod := range podList {
			Expect(pod.Spec.NodeName).To(Equal(nodeOne))
		}

		By("Remove the node label")
		removeLabelOffNode(context, nodeOne, []string{k})
		err = verifyLabelsRemoved(context, nodeOne, []string{k})
		Expect(err).NotTo(HaveOccurred())
	})

	It("validates that a Job with requiredDuringSchedulingIgnoredDuringExecution stays pending if no node match the constraint [Hard NodeAffinity]", func() {
		context := initTestContext()
		defer cleanupTestContext(context)

		By("Create a Job with requiredDuringSchedulingIgnoredDuringExecution constraint but with node with labels")
		nodeAffinityHard := &v1.Affinity{
			NodeAffinity: &v1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
					NodeSelectorTerms: []v1.NodeSelectorTerm{
						{
							MatchExpressions: []v1.NodeSelectorRequirement{
								{
									Key:      "gpu",
									Operator: v1.NodeSelectorOpIn,
									Values: []string{
										"nvidia",
										"zotac-nvidia",
									},
								},
							},
						},
					},
				},
			},
		}
		By("Trying to launch a Job, with requiredDuringSchedulingIgnoredDuringExecution constraint on its tasks")
		nodeAffinityHardjs := &jobSpec{
			name: "node-affinity-hard",
			tasks: []taskSpec{
				{
					img:      "nginx",
					min:      1,
					rep:      1,
					affinity: nodeAffinityHard,
				},
			},
		}
		_, nodeAffinityHardpg := createJob(context, nodeAffinityHardjs)
		By("Validate if the Job is not running in any node")
		err := waitPodGroupReady(context, nodeAffinityHardpg)
		// This will timeout waiting for the Podgroup to be ready
		if err != wait.ErrWaitTimeout {
			checkError(context, err)
		}
	})
})
