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

package util

import (
	"context"
	"math/rand"
	"sort"
	"sync"

	"github.com/golang/glog"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/priorities"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"github.com/kubernetes-sigs/kube-batch/pkg/scheduler/api"
)

// HostPriority represents the priority of scheduling to a particular host, higher priority is better.
type HostPriority struct {
	// Name of the host
	Host string
	// Score associated with the host
	Score float64
}

// HostPriorityList declares a []HostPriority type.
type HostPriorityList []HostPriority

func (h HostPriorityList) Len() int {
	return len(h)
}

func (h HostPriorityList) Less(i, j int) bool {
	if h[i].Score == h[j].Score {
		return h[i].Host < h[j].Host
	}
	return h[i].Score < h[j].Score
}

func (h HostPriorityList) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

// PredicateNodes returns nodes that fit task
func PredicateNodes(task *api.TaskInfo, nodes []*api.NodeInfo, fn api.PredicateFn) []*api.NodeInfo {
	var predicateNodes []*api.NodeInfo

	var workerLock sync.Mutex
	checkNode := func(index int) {
		node := nodes[index]
		glog.V(3).Infof("Considering Task <%v/%v> on node <%v>: <%v> vs. <%v>",
			task.Namespace, task.Name, node.Name, task.Resreq, node.Idle)

		// TODO (k82cn): Enable eCache for performance improvement.
		if err := fn(task, node); err != nil {
			glog.Errorf("Predicates failed for task <%s/%s> on node <%s>: %v",
				task.Namespace, task.Name, node.Name, err)
			return
		}

		workerLock.Lock()
		predicateNodes = append(predicateNodes, node)
		workerLock.Unlock()
	}

	workqueue.ParallelizeUntil(context.TODO(), 16, len(nodes), checkNode)
	return predicateNodes
}

// PrioritizeNodes returns a map whose key is node's score and value are corresponding nodes
func PrioritizeNodes(
	task *api.TaskInfo,
	filterNodes []*api.NodeInfo,
	priorityConfigs []priorities.PriorityConfig,
) (HostPriorityList, error) {
	nodeNameToInfo, nodes := generateNodeMapAndSlice(filterNodes)
	var (
		mu   = sync.Mutex{}
		wg   = sync.WaitGroup{}
		errs []error
	)
	appendError := func(err error) {
		mu.Lock()
		defer mu.Unlock()
		errs = append(errs, err)
	}

	results := make([]schedulerapi.HostPriorityList, len(priorityConfigs), len(priorityConfigs))

	for i, priorityConfig := range priorityConfigs {
		if priorityConfig.Function != nil {
			wg.Add(1)
			go func(index int, config priorities.PriorityConfig) {
				defer wg.Done()
				var err error
				results[index], err = config.Function(task.Pod, nodeNameToInfo, nodes)
				if err != nil {
					appendError(err)
				}
			}(i, priorityConfig)
		} else {
			results[i] = make(schedulerapi.HostPriorityList, len(nodes))
		}
	}
	processNode := func(index int) {
		nodeInfo := nodeNameToInfo[nodes[index].Name]
		var err error
		for i := range priorityConfigs {
			if priorityConfigs[i].Function != nil {
				continue
			}
			results[i][index], err = priorityConfigs[i].Map(task.Pod, nil, nodeInfo)
			if err != nil {
				appendError(err)
				return
			}
		}
	}
	workqueue.ParallelizeUntil(context.TODO(), 16, len(nodes), processNode)
	for i, priorityConfig := range priorityConfigs {
		if priorityConfig.Reduce == nil {
			continue
		}
		wg.Add(1)
		go func(index int, config priorities.PriorityConfig) {
			defer wg.Done()
			if err := config.Reduce(task.Pod, nil, nodeNameToInfo, results[index]); err != nil {
				appendError(err)
			}
			if glog.V(10) {
				for _, hostPriority := range results[index] {
					glog.Infof("%v/%v -> %v: %v, Score: (%d)", task.Pod.Namespace, task.Pod.Name, hostPriority.Host, config.Name, hostPriority.Score)
				}
			}
		}(i, priorityConfig)
	}
	// Wait for all computations to be finished.
	wg.Wait()
	if len(errs) != 0 {
		return HostPriorityList{}, errors.NewAggregate(errs)
	}

	// Summarize all scores.
	result := make(HostPriorityList, 0, len(nodes))
	for i := range nodes {
		result = append(result, HostPriority{Host: nodes[i].Name, Score: 0})
		for j := range priorityConfigs {
			result[i].Score += float64(results[j][i].Score * priorityConfigs[j].Weight)
		}
	}

	return result, nil
}

// SortNodes returns nodes by order of score
func SortNodes(priorityList HostPriorityList, nodesInfo map[string]*api.NodeInfo) []*api.NodeInfo {
	var nodesInorder []*api.NodeInfo

	sort.Sort(sort.Reverse(priorityList))

	for _, hostPriority := range priorityList {
		node := nodesInfo[hostPriority.Host]
		nodesInorder = append(nodesInorder, node)
	}

	return nodesInorder
}

// SelectBestNode returns best node whose score is highest, pick one randomly if there are many nodes with same score.
func SelectBestNode(priorityList HostPriorityList) string {
	maxScores := findMaxScores(priorityList)
	ix := rand.Intn(len(maxScores))
	return priorityList[maxScores[ix]].Host
}

// findMaxScores returns the indexes of nodes in the "priorityList" that has the highest "Score".
func findMaxScores(priorityList HostPriorityList) []int {
	maxScoreIndexes := make([]int, 0, len(priorityList)/2)
	maxScore := priorityList[0].Score
	for i, hp := range priorityList {
		if hp.Score > maxScore {
			maxScore = hp.Score
			maxScoreIndexes = maxScoreIndexes[:0]
			maxScoreIndexes = append(maxScoreIndexes, i)
		} else if hp.Score == maxScore {
			maxScoreIndexes = append(maxScoreIndexes, i)
		}
	}
	return maxScoreIndexes
}

// GetNodeList returns values of the map 'nodes'
func GetNodeList(nodes map[string]*api.NodeInfo) []*api.NodeInfo {
	result := make([]*api.NodeInfo, 0, len(nodes))
	for _, v := range nodes {
		result = append(result, v)
	}
	return result
}

func generateNodeMapAndSlice(nodes []*api.NodeInfo) (map[string]*schedulernodeinfo.NodeInfo, []*v1.Node) {
	var nodeMap map[string]*schedulernodeinfo.NodeInfo
	var nodeSlice []*v1.Node
	nodeMap = make(map[string]*schedulernodeinfo.NodeInfo)
	for _, node := range nodes {
		nodeInfo := schedulernodeinfo.NewNodeInfo(node.Pods()...)
		nodeInfo.SetNode(node.Node)
		nodeMap[node.Name] = nodeInfo
		nodeSlice = append(nodeSlice, node.Node)
	}
	return nodeMap, nodeSlice
}
