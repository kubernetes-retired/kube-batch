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
	"github.com/kubernetes-sigs/kube-batch/pkg/scheduler/plugins/util"

	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/priorities"
)

const (
	// NodeAffinityWeight is the key for providing Node Affinity Priority Weight in YAML
	NodeAffinityWeight = "nodeaffinity.weight"
	// PodAffinityWeight is the key for providing Pod Affinity Priority Weight in YAML
	PodAffinityWeight = "podaffinity.weight"
	// LeastRequestedWeight is the key for providing Least Requested Priority Weight in YAML
	LeastRequestedWeight = "leastrequested.weight"
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

//New function returns prioritizePlugin object
func New(aruguments framework.Arguments) framework.Plugin {
	return &nodeOrderPlugin{pluginArguments: aruguments}
}

func (pp *nodeOrderPlugin) Name() string {
	return "nodeorder"
}

type priorityWeight struct {
	leastReqWeight          int
	nodeAffinityWeight      int
	podAffinityWeight       int
	balancedRescourceWeight int
}

func calculateWeight(args framework.Arguments) priorityWeight {
	/*
	   User Should give priorityWeight in this format(nodeaffinity.weight, podaffinity.weight, leastrequested.weight, balancedresource.weight).
	   Currently supported only for nodeaffinity, podaffinity, leastrequested, balancedresouce priorities.

	   actions: "reclaim, allocate, backfill, preempt"
	   tiers:
	   - plugins:
	     - name: priority
	     - name: gang
	     - name: conformance
	   - plugins:
	     - name: drf
	     - name: predicates
	     - name: proportion
	     - name: nodeorder
	       arguments:
	         nodeaffinity.weight: 2
	         podaffinity.weight: 2
	         leastrequested.weight: 2
	         balancedresource.weight: 2
	*/

	// Values are initialized to 1.
	weight := priorityWeight{
		leastReqWeight:          1,
		nodeAffinityWeight:      1,
		podAffinityWeight:       1,
		balancedRescourceWeight: 1,
	}

	// Checks whether nodeaffinity.weight is provided or not, if given, modifies the value in weight struct.
	args.GetInt(&weight.nodeAffinityWeight, NodeAffinityWeight)

	// Checks whether podaffinity.weight is provided or not, if given, modifies the value in weight struct.
	args.GetInt(&weight.podAffinityWeight, PodAffinityWeight)

	// Checks whether leastrequested.weight is provided or not, if given, modifies the value in weight struct.
	args.GetInt(&weight.leastReqWeight, LeastRequestedWeight)

	// Checks whether balancedresource.weight is provided or not, if given, modifies the value in weight struct.
	args.GetInt(&weight.balancedRescourceWeight, BalancedResourceWeight)

	return weight
}

func (pp *nodeOrderPlugin) OnSessionOpen(ssn *framework.Session) {
	weight := calculateWeight(pp.pluginArguments)

	pl := &util.PodLister{
		Session: ssn,
	}

	nl := &util.NodeLister{
		Session: ssn,
	}

	cn := &cachedNodeInfo{
		session: ssn,
	}

	priorityConfigs := []algorithm.PriorityConfig{
		{
			Name:   "LeastRequestedPriority",
			Map:    priorities.LeastRequestedPriorityMap,
			Weight: weight.leastReqWeight,
		},
		{
			Name:   "BalancedResourceAllocation",
			Map:    priorities.BalancedResourceAllocationMap,
			Weight: weight.balancedRescourceWeight,
		},
		{
			Name:   "NodeAffinityPriority",
			Map:    priorities.CalculateNodeAffinityPriorityMap,
			Reduce: priorities.CalculateNodeAffinityPriorityReduce,
			Weight: weight.balancedRescourceWeight,
		},
		{
			Name:     "InterPodAffinityPriority",
			Function: priorities.NewInterPodAffinityPriority(cn, nl, pl, v1.DefaultHardPodAffinitySymmetricWeight),
			Weight:   weight.balancedRescourceWeight,
		},
	}
	ssn.AddNodePrioritizers(pp.Name(), priorityConfigs)
}

func (pp *nodeOrderPlugin) OnSessionClose(ssn *framework.Session) {
}
