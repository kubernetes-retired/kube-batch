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

package nodeorder

import (
	"fmt"

	"github.com/kubernetes-sigs/kube-batch/pkg/scheduler/framework"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/priorities"
)

const (
	// LeastRequestedWeight is the key for providing Least Requested Priority Weight in YAML
	LeastRequestedWeight = "leastrequested.weight"
	// MostRequestedWeight is the key for providing Most Requested Priority Weight in YAML
	MostRequestedWeight = "mostrequested.weight"
	// NodeAffinityWeight is the key for providing Node Affinity Priority Weight in YAML
	NodeAffinityWeight = "nodeaffinity.weight"
	// PodAffinityWeight is the key for providing Pod Affinity Priority Weight in YAML
	PodAffinityWeight = "podaffinity.weight"
	// BalancedResourceWeight is the key for providing Balanced Resource Priority Weight in YAML
	BalancedResourceWeight = "balancedresource.weight"
)

type nodeOrderPlugin struct {
	// Arguments given for the plugin
	pluginArguments framework.Arguments
}

type cachedNodeInfo struct {
	session *framework.Session
}

func (c *cachedNodeInfo) GetNodeInfo(name string) (*v1.Node, error) {
	node, found := c.session.Nodes[name]
	if !found {
		for _, cacheNode := range c.session.Nodes {
			pods := cacheNode.Pods()
			for _, pod := range pods {
				if pod.Spec.NodeName == "" {
					return cacheNode.Node, nil
				}
			}
		}
		return nil, fmt.Errorf("failed to find node <%s>", name)
	}

	return node.Node, nil
}

// New function returns nodeorder Plugin object.
func New(aruguments framework.Arguments) framework.Plugin {
	return &nodeOrderPlugin{pluginArguments: aruguments}
}

func (pp *nodeOrderPlugin) Name() string {
	return "nodeorder"
}

type priorityWeight struct {
	leastReqWeight         int
	mostReqWeight          int
	nodeAffinityWeight     int
	podAffinityWeight      int
	balancedResourceWeight int
}

// calculateWeight from the provided arguments.
//
// Currently only supported priorities are nodeaffinity, podaffinity, leastrequested,
// mostrequested, balancedresouce.
//
// User should specify priority weights in the config in this format:
//
//  actions: "reclaim, allocate, backfill, preempt"
//  tiers:
//  - plugins:
//    - name: priority
//    - name: gang
//    - name: conformance
//  - plugins:
//    - name: drf
//    - name: predicates
//    - name: proportion
//    - name: nodeorder
//      arguments:
//        leastrequested.weight: 2
//        mostrequested.weight: 0
//        nodeaffinity.weight: 2
//        podaffinity.weight: 2
//        balancedresource.weight: 2
func calculateWeight(args framework.Arguments) priorityWeight {
	// Initial values for weights.
	// By default, for backward compatibility and for reasonable scores,
	// least requested priority is enabled and most requested priority is disabled.
	weight := priorityWeight{
		leastReqWeight:         1,
		mostReqWeight:          0,
		nodeAffinityWeight:     1,
		podAffinityWeight:      1,
		balancedResourceWeight: 1,
	}

	// Checks whether leastrequested.weight is provided or not, if given, modifies the value in weight struct.
	args.GetInt(&weight.leastReqWeight, LeastRequestedWeight)
	// Checks whether mostrequested.weight is provided or not, if given, modifies the value in weight struct.
	args.GetInt(&weight.mostReqWeight, MostRequestedWeight)
	// Checks whether nodeaffinity.weight is provided or not, if given, modifies the value in weight struct.
	args.GetInt(&weight.nodeAffinityWeight, NodeAffinityWeight)
	// Checks whether podaffinity.weight is provided or not, if given, modifies the value in weight struct.
	args.GetInt(&weight.podAffinityWeight, PodAffinityWeight)
	// Checks whether balancedresource.weight is provided or not, if given, modifies the value in weight struct.
	args.GetInt(&weight.balancedResourceWeight, BalancedResourceWeight)

	return weight
}

func (pp *nodeOrderPlugin) OnSessionOpen(ssn *framework.Session) {
	weight := calculateWeight(pp.pluginArguments)

	cn := &cachedNodeInfo{
		session: ssn,
	}

	priorityConfigs := []priorities.PriorityConfig{
		{
			Name:   "LeastRequestedPriority",
			Map:    priorities.LeastRequestedPriorityMap,
			Weight: weight.leastReqWeight,
		},
		{
			Name:   "MostRequestedPriority",
			Map:    priorities.MostRequestedPriorityMap,
			Weight: weight.mostReqWeight,
		},
		{
			Name:   "NodeAffinityPriority",
			Map:    priorities.CalculateNodeAffinityPriorityMap,
			Reduce: priorities.CalculateNodeAffinityPriorityReduce,
			Weight: weight.nodeAffinityWeight,
		},
		{
			Name:     "InterPodAffinityPriority",
			Function: priorities.NewInterPodAffinityPriority(cn, v1.DefaultHardPodAffinitySymmetricWeight),
			Weight:   weight.podAffinityWeight,
		},
		{
			Name:   "BalancedResourceAllocation",
			Map:    priorities.BalancedResourceAllocationMap,
			Weight: weight.balancedResourceWeight,
		},
	}
	ssn.AddNodePrioritizers(pp.Name(), priorityConfigs)
}

func (pp *nodeOrderPlugin) OnSessionClose(ssn *framework.Session) {
}
