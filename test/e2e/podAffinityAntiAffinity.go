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
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Pod Affinity and Anti-Affinity", func() {

	// We have two setup podgroups which will create a single task with single pod.
	// This is required to test the Pod Affinity and Anti-Affinity.

	// since these setup pods will be created ramdomly
	// we store their node name in these var
	var setupPodGrpOnesNodeName, setupPodGrpTwosNodeName string
	setupPodGrpOneName := "setup-pod-one"
	setupPodGrpTwoName := "setup-pod-two"
	testPodGrpName := "test-pod"

	// we create and delete these podgroups in the following sub-specs
	setupPodGrpOne := &jobSpec{
		name: setupPodGrpOneName,
		tasks: []taskSpec{
			{
				img: "nginx",
				min: 1,
				rep: 1,
				labels: map[string]string{
					"security": "S1",
				},
			},
		},
	}

	setupPodGrpTwo := &jobSpec{
		name: setupPodGrpTwoName,
		tasks: []taskSpec{
			{
				img: "nginx",
				min: 1,
				rep: 1,
				labels: map[string]string{
					"security": "S2",
				},
			},
		},
	}

	//1
	It("validate scheduler respect's pod-affinity with hard constraint", func() {
		context := initTestContext()
		defer cleanupTestContext(context)

		podAffinityHard := &v1.Affinity{
			PodAffinity: &v1.PodAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "security",
									Operator: metav1.LabelSelectorOpIn,
									Values: []string{
										"S1",
									},
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		}

		testPodGrp := &jobSpec{
			name: testPodGrpName,
			tasks: []taskSpec{
				{
					img:      "nginx",
					min:      1,
					rep:      1,
					affinity: podAffinityHard,
				},
			},
		}

		By("Trying to launch the setup podgroup one with label security S1")
		_, setupPgOne := createJob(context, setupPodGrpOne)

		By("validate if setup podgroup is running and get the node of the setup podgroup")
		err := waitPodGroupReady(context, setupPgOne)
		checkError(context, err)

		nodenames := getNodesOfPodGroup(context, setupPgOne)

		// this should be a single nodename
		By("Check if the Podgroup is running on a single node, since it has only one task")
		Expect(len(nodenames)).To(Equal(int(1)))

		By(fmt.Sprintf("Setup PodGroup is running on %s", nodenames[0]))
		setupPodGrpOnesNodeName = nodenames[0]

		By("Deploy the test PodGroup with pod affinity hard constraint and check if its running")
		_, testPg := createJob(context, testPodGrp)
		err = waitPodGroupReady(context, testPg)
		checkError(context, err)

		By("Get the node name of the Podgroup")
		nodenames = getNodesOfPodGroup(context, testPg)

		// this should be a single nodename
		By("Check if the Podgroup is running on a single node, since it has only one task")
		Expect(len(nodenames)).To(Equal(int(1)))

		By("validate if test PodGroup is running on the right node")
		Expect(nodenames[0]).To(Equal(setupPodGrpOnesNodeName))
	})

	//2
	It("validates scheduler respect's pod-anti-affinity with hard constraint", func() {
		context := initTestContext()
		defer cleanupTestContext(context)

		podAntiAffinityHard := &v1.Affinity{
			PodAntiAffinity: &v1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "security",
									Operator: metav1.LabelSelectorOpIn,
									Values: []string{
										"S1",
									},
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		}

		testPodGrp := &jobSpec{
			name: testPodGrpName,
			tasks: []taskSpec{
				{
					img:      "nginx",
					min:      1,
					rep:      1,
					affinity: podAntiAffinityHard,
				},
			},
		}

		By("Trying to launch the setup podgroup one with label security S1")
		_, setupPgOne := createJob(context, setupPodGrpOne)

		By("validate if setup podgroup is running and get the node of the setup podgroup")
		err := waitPodGroupReady(context, setupPgOne)
		checkError(context, err)

		nodeNames := getNodesOfPodGroup(context, setupPgOne)

		// this should be a single node name
		By("Check if the Podgroup is running on a single node, since it has only one task")
		Expect(len(nodeNames)).To(Equal(int(1)))

		By(fmt.Sprintf("Setup PodGroup is running on %s", nodeNames[0]))
		setupPodGrpOnesNodeName = nodeNames[0]

		By("Deploy the test PodGroup with pod affinity hard constraint and check if its running")
		_, testPg := createJob(context, testPodGrp)
		err = waitPodGroupReady(context, testPg)
		checkError(context, err)

		By("Get the node name of the Podgroup")
		nodeNames = getNodesOfPodGroup(context, testPg)

		// this should be a single nodename
		By("Check if the Podgroup is running on a single node, since it has only one task")
		Expect(len(nodeNames)).To(Equal(int(1)))

		By("validate if test PodGroup is running on the right node")
		Expect(nodeNames[0]).NotTo(Equal(setupPodGrpOnesNodeName))
	})

	//3
	It("validates scheduler respect's pod-anti-affinity with soft constraint", func() {
		context := initTestContext()
		defer cleanupTestContext(context)

		podAntiAffinitySoft := &v1.Affinity{
			PodAntiAffinity: &v1.PodAntiAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
					{
						Weight: 100,
						PodAffinityTerm: v1.PodAffinityTerm{
							LabelSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      "security",
										Operator: metav1.LabelSelectorOpIn,
										Values: []string{
											"S1",
										},
									},
								},
							},
							TopologyKey: "kubernetes.io/hostname",
						},
					},
				},
			},
		}

		testPodGrp := &jobSpec{
			name: testPodGrpName,
			tasks: []taskSpec{
				{
					img:      "nginx",
					min:      1,
					rep:      1,
					affinity: podAntiAffinitySoft,
				},
			},
		}

		By("Trying to launch the setup podgroup one with label security S1")
		_, setupPgOne := createJob(context, setupPodGrpOne)

		By("validate if setup podgroup is running and get the node of the setup podgroup")
		err := waitPodGroupReady(context, setupPgOne)
		checkError(context, err)

		nodeNames := getNodesOfPodGroup(context, setupPgOne)

		// this should be a single node name
		By("Check if the Podgroup is running on a single node, since it has only one task")
		Expect(len(nodeNames)).To(Equal(int(1)))

		By(fmt.Sprintf("Setup PodGroup is running on %s", nodeNames[0]))
		setupPodGrpOnesNodeName = nodeNames[0]

		By("Deploy the test PodGroup with pod affinity hard constraint and check if its running")
		_, testPg := createJob(context, testPodGrp)
		err = waitPodGroupReady(context, testPg)
		checkError(context, err)

		By("Get the node name of the Podgroup")
		nodeNames = getNodesOfPodGroup(context, testPg)

		// this should be a single nodename
		By("Check if the Podgroup is running on a single node, since it has only one task")
		Expect(len(nodeNames)).To(Equal(int(1)))

		By("validate if test PodGroup is running on the right node")
		Expect(nodeNames[0]).NotTo(Equal(setupPodGrpOnesNodeName))
	})

	//4
	It("validates scheduler respect's a pod-affinity with both hard and soft constraint", func() {
		context := initTestContext()
		defer cleanupTestContext(context)

		podAffinityHardSoft := &v1.Affinity{
			PodAffinity: &v1.PodAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "security",
									Operator: metav1.LabelSelectorOpIn,
									Values: []string{
										"S1",
									},
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
				PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
					{
						Weight: 100,
						PodAffinityTerm: v1.PodAffinityTerm{
							LabelSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      "security",
										Operator: metav1.LabelSelectorOpIn,
										Values: []string{
											"S2",
										},
									},
								},
							},
							TopologyKey: "kubernetes.io/hostname",
						},
					},
				},
			},
		}

		testPodGrp := &jobSpec{
			name: testPodGrpName,
			tasks: []taskSpec{
				{
					img:      "nginx",
					min:      1,
					rep:      1,
					affinity: podAffinityHardSoft,
				},
			},
		}

		By("Trying to launch the setup podgroup one with label security S1")
		_, setupPgOne := createJob(context, setupPodGrpOne)

		By("validate if setup podgroup is running and get the node of the setup podgroup")
		err := waitPodGroupReady(context, setupPgOne)
		checkError(context, err)

		nodeNames := getNodesOfPodGroup(context, setupPgOne)

		// this should be a single node name
		By("Check if the Podgroup is running on a single node, since it has only one task")
		Expect(len(nodeNames)).To(Equal(int(1)))

		By(fmt.Sprintf("Setup PodGroup is running on %s", nodeNames[0]))
		setupPodGrpOnesNodeName = nodeNames[0]

		// We deploy the second setup podgroup here for this test case
		By("Trying to launch the setup podgroup two with label security S2")
		_, setupPgTwo := createJob(context, setupPodGrpTwo)

		By("validate if setup podgroup two is running and get the node of the setup podgroup")
		err = waitPodGroupReady(context, setupPgTwo)
		checkError(context, err)

		nodeNames = getNodesOfPodGroup(context, setupPgTwo)

		// this should be a single node name
		By("Check if the Podgroup is running on a single node, since it has only one task")
		Expect(len(nodeNames)).To(Equal(int(1)))

		By(fmt.Sprintf("Setup PodGroup is running on %s", nodeNames[0]))
		setupPodGrpTwosNodeName = nodeNames[0]

		By("Deploy the test PodGroup with pod affinity hard constraint and check if its running")
		_, testPg := createJob(context, testPodGrp)
		err = waitPodGroupReady(context, testPg)
		checkError(context, err)

		By("Get the node name of the test Podgroup")
		nodeNames = getNodesOfPodGroup(context, testPg)

		// this should be a single nodename
		By("Check if the Podgroup is running on a single node, since it has only one task")
		Expect(len(nodeNames)).To(Equal(int(1)))

		By("validate if test PodGroup is running on the right node")
		Expect(nodeNames[0]).To(Equal(setupPodGrpOnesNodeName))

	})

	//5
	It("validates scheduler respect's a pod-anti-affinity with both hard and soft constraint", func() {
		context := initTestContext()
		defer cleanupTestContext(context)

		podAntiAffinityHardSoft := &v1.Affinity{
			PodAntiAffinity: &v1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "security",
									Operator: metav1.LabelSelectorOpIn,
									Values: []string{
										"S1",
									},
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
				PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
					{
						Weight: 100,
						PodAffinityTerm: v1.PodAffinityTerm{
							LabelSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      "security",
										Operator: metav1.LabelSelectorOpIn,
										Values: []string{
											"S2",
										},
									},
								},
							},
							TopologyKey: "kubernetes.io/hostname",
						},
					},
				},
			},
		}

		testPodGrp := &jobSpec{
			name: testPodGrpName,
			tasks: []taskSpec{
				{
					img:      "nginx",
					min:      1,
					rep:      1,
					affinity: podAntiAffinityHardSoft,
				},
			},
		}

		By("Trying to launch the setup podgroup one with label security S1")
		_, setupPgOne := createJob(context, setupPodGrpOne)

		By("validate if setup podgroup is running and get the node of the setup podgroup")
		err := waitPodGroupReady(context, setupPgOne)
		checkError(context, err)

		nodeNames := getNodesOfPodGroup(context, setupPgOne)

		// this should be a single node name
		By("Check if the Podgroup is running on a single node, since it has only one task")
		Expect(len(nodeNames)).To(Equal(int(1)))

		By(fmt.Sprintf("Setup PodGroup is running on %s", nodeNames[0]))
		setupPodGrpOnesNodeName = nodeNames[0]

		// We deploy the second setup podgroup here for this test case
		By("Trying to launch the setup podgroup two with label security S2")
		_, setupPgTwo := createJob(context, setupPodGrpTwo)

		By("validate if setup podgroup two is running and get the node of the setup podgroup")
		err = waitPodGroupReady(context, setupPgTwo)
		checkError(context, err)

		nodeNames = getNodesOfPodGroup(context, setupPgTwo)

		// this should be a single node name
		By("Check if the Podgroup is running on a single node, since it has only one task")
		Expect(len(nodeNames)).To(Equal(int(1)))

		By(fmt.Sprintf("Setup PodGroup is running on %s", nodeNames[0]))
		setupPodGrpTwosNodeName = nodeNames[0]

		By("Deploy the test PodGroup with pod affinity hard constraint and check if its running")
		_, testPg := createJob(context, testPodGrp)
		err = waitPodGroupReady(context, testPg)
		checkError(context, err)

		By("Get the node name of the test Podgroup")
		nodeNames = getNodesOfPodGroup(context, testPg)

		// this should be a single nodename
		By("Check if the Podgroup is running on a single node, since it has only one task")
		Expect(len(nodeNames)).To(Equal(int(1)))

		By("validate if test PodGroup is running on the right node")
		Expect(nodeNames[0]).ShouldNot(SatisfyAll(Equal(setupPodGrpOnesNodeName), Equal(setupPodGrpTwosNodeName)))
	})

	//6
	It("validates scheduler respect's a pod-affinity with hard constraint and anti-affinity soft constraint", func() {
		context := initTestContext()
		defer cleanupTestContext(context)

		podAntiAffinityHardSoft := &v1.Affinity{
			PodAffinity: &v1.PodAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "security",
									Operator: metav1.LabelSelectorOpIn,
									Values: []string{
										"S1",
									},
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
			PodAntiAffinity: &v1.PodAntiAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
					{
						Weight: 100,
						PodAffinityTerm: v1.PodAffinityTerm{
							LabelSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      "security",
										Operator: metav1.LabelSelectorOpIn,
										Values: []string{
											"S2",
										},
									},
								},
							},
							TopologyKey: "kubernetes.io/hostname",
						},
					},
				},
			},
		}

		testPodGrp := &jobSpec{
			name: testPodGrpName,
			tasks: []taskSpec{
				{
					img:      "nginx",
					min:      1,
					rep:      1,
					affinity: podAntiAffinityHardSoft,
				},
			},
		}

		By("Trying to launch the setup podgroup one with label security S1")
		_, setupPgOne := createJob(context, setupPodGrpOne)

		By("validate if setup podgroup is running and get the node of the setup podgroup")
		err := waitPodGroupReady(context, setupPgOne)
		checkError(context, err)

		nodeNames := getNodesOfPodGroup(context, setupPgOne)

		// this should be a single node name
		By("Check if the Podgroup is running on a single node, since it has only one task")
		Expect(len(nodeNames)).To(Equal(int(1)))

		By(fmt.Sprintf("Setup PodGroup is running on %s", nodeNames[0]))
		setupPodGrpOnesNodeName = nodeNames[0]

		// We deploy the second setup podgroup here for this test case
		By("Trying to launch the setup podgroup two with label security S2")
		_, setupPgTwo := createJob(context, setupPodGrpTwo)

		By("validate if setup podgroup two is running and get the node of the setup podgroup")
		err = waitPodGroupReady(context, setupPgTwo)
		checkError(context, err)

		nodeNames = getNodesOfPodGroup(context, setupPgTwo)

		// this should be a single node name
		By("Check if the Podgroup is running on a single node, since it has only one task")
		Expect(len(nodeNames)).To(Equal(int(1)))

		By(fmt.Sprintf("Setup PodGroup is running on %s", nodeNames[0]))
		setupPodGrpTwosNodeName = nodeNames[0]

		By("Deploy the test PodGroup with pod affinity hard constraint and check if its running")
		_, testPg := createJob(context, testPodGrp)
		err = waitPodGroupReady(context, testPg)
		checkError(context, err)

		By("Get the node name of the test Podgroup")
		nodeNames = getNodesOfPodGroup(context, testPg)

		// this should be a single nodename
		By("Check if the Podgroup is running on a single node, since it has only one task")
		Expect(len(nodeNames)).To(Equal(int(1)))

		By("validate if test PodGroup is running on the right node")
		Expect(nodeNames[0]).To(Equal(setupPodGrpOnesNodeName))
	})

	//7
	It("validates scheduler respect's a pod-anti-affinity with hard constraint and affinity soft constraint", func() {
		context := initTestContext()
		defer cleanupTestContext(context)

		podAntiAffinityHardSoft := &v1.Affinity{
			PodAntiAffinity: &v1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "security",
									Operator: metav1.LabelSelectorOpIn,
									Values: []string{
										"S1",
									},
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
			PodAffinity: &v1.PodAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
					{
						Weight: 100,
						PodAffinityTerm: v1.PodAffinityTerm{
							LabelSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      "security",
										Operator: metav1.LabelSelectorOpIn,
										Values: []string{
											"S2",
										},
									},
								},
							},
							TopologyKey: "kubernetes.io/hostname",
						},
					},
				},
			},
		}

		testPodGrp := &jobSpec{
			name: testPodGrpName,
			tasks: []taskSpec{
				{
					img:      "nginx",
					min:      1,
					rep:      1,
					affinity: podAntiAffinityHardSoft,
				},
			},
		}

		By("Trying to launch the setup podgroup one with label security S1")
		_, setupPgOne := createJob(context, setupPodGrpOne)

		By("validate if setup podgroup is running and get the node of the setup podgroup")
		err := waitPodGroupReady(context, setupPgOne)
		checkError(context, err)

		nodeNames := getNodesOfPodGroup(context, setupPgOne)

		// this should be a single node name
		By("Check if the Podgroup is running on a single node, since it has only one task")
		Expect(len(nodeNames)).To(Equal(int(1)))

		By(fmt.Sprintf("Setup PodGroup is running on %s", nodeNames[0]))
		setupPodGrpOnesNodeName = nodeNames[0]

		// We deploy the second setup podgroup here for this test case
		By("Trying to launch the setup podgroup two with label security S2")
		_, setupPgTwo := createJob(context, setupPodGrpTwo)

		By("validate if setup podgroup two is running and get the node of the setup podgroup")
		err = waitPodGroupReady(context, setupPgTwo)
		checkError(context, err)

		nodeNames = getNodesOfPodGroup(context, setupPgTwo)

		// this should be a single node name
		By("Check if the Podgroup is running on a single node, since it has only one task")
		Expect(len(nodeNames)).To(Equal(int(1)))

		By(fmt.Sprintf("Setup PodGroup is running on %s", nodeNames[0]))
		setupPodGrpTwosNodeName = nodeNames[0]

		By("Deploy the test PodGroup with pod affinity hard constraint and check if its running")
		_, testPg := createJob(context, testPodGrp)
		err = waitPodGroupReady(context, testPg)
		checkError(context, err)

		By("Get the node name of the test Podgroup")
		nodeNames = getNodesOfPodGroup(context, testPg)

		// this should be a single nodename
		By("Check if the Podgroup is running on a single node, since it has only one task")
		Expect(len(nodeNames)).To(Equal(int(1)))

		By("validate if test PodGroup is running on the right node")
		Expect(nodeNames[0]).NotTo(Equal(setupPodGrpOnesNodeName))

	})
})
