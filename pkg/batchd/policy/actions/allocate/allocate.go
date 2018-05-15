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

package allocate

import (
	"github.com/golang/glog"

	policyapi "github.com/kubernetes-incubator/kube-arbitrator/pkg/batchd/policy/api"
	"github.com/kubernetes-incubator/kube-arbitrator/pkg/batchd/policy/framework"
	"github.com/kubernetes-incubator/kube-arbitrator/pkg/batchd/policy/util"
)

type allocateAction struct {
}

func New() *allocateAction {
	return &allocateAction{}
}

func (drf *allocateAction) Name() string {
	return "allocate"
}

func (drf *allocateAction) Initialize() {}

func (drf *allocateAction) Execute(ssn *framework.Session) {
	glog.V(4).Infof("Enter Execute ...")
	defer glog.V(4).Infof("Leaving Execute ...")

	requests := util.NewPriorityQueue(ssn.RequestOrderFn)
	pendings := map[string]*util.PriorityQueue{}

	for _, req := range ssn.Requests {
		requests.Push(req)

		ps := util.NewPriorityQueue(nil)
		for _, ru := range req.Units {
			if ru.Status == policyapi.Pending {
				ps.Push(ru)
			}
		}

		pendings[req.ID] = ps
	}

	for !requests.Empty() {
		req := requests.Pop().(*policyapi.Request)

		glog.V(3).Infof("Try to allocate resources to Request <%v>", req.ID)

		rus := pendings[req.ID]

		p := rus.Pop().(*policyapi.RequestUnit)
		if p == nil {
			glog.V(3).Infof("no pending units in Request <%v/%v>")
			continue
		}

		assigned := false
		for _, n := range ssn.Nodes {
			if p.Resreq.Less(n.Idle) {
				ssn.PublishDecision(&policyapi.Decision{
					RequestID:     req.ID,
					RequestUnitID: p.ID,
					Host:          n.Name,
					Type:          policyapi.Allocate,
				})
				assigned = true

				break
			}
		}

		if assigned {
			// push Request back for next allocation
			requests.Push(req)
		} else {
			glog.V(3).Infof("Failed to allocate resources to Request <%v>", req.ID)
		}
	}

	// TODO: move this to 'predicate'
	//pq := util.NewPriorityQueue()
	//matchNodesForPodSet := make(map[string][]*policyapi.Node)

	//for _, q := range dq {
	//	ru := q.Value.(*policyapi.RequestUnit)
	//
	//	// fetch the nodes that match PodSet NodeSelector and NodeAffinity
	//	// and store it for following DRF assignment
	//	//matchNodes := fetchMatchNodeForPodSet(psi, ssn.Nodes)
	//	//matchNodesForPodSet[psi.podSet.Name] = matchNodes
	//
	//	assigned := drf.assignMinimalPods(ru.insufficientMinAvailable(), ru)
	//	if assigned {
	//		// only push PodSet with MinAvailable to priority queue
	//		// to avoid PodSet get resources less than MinAvailable by following DRF assignment
	//		pq.Push(psi, psi.share)
	//
	//		glog.V(3).Infof("assign MinAvailable for podset %s/%s successfully",
	//			psi.podSet.Namespace, psi.podSet.Name)
	//	} else {
	//		glog.V(3).Infof("assign MinAvailable for podset %s/%s failed, there is no enough resources",
	//			psi.podSet.Namespace, psi.podSet.Name)
	//	}
	//}

}

func (drf *allocateAction) UnInitialize() {}

//
//func (drf *allocateAction) allocate(ssn *framework.Session, reqID string, rus *util.PriorityQueue) bool {
//	glog.V(4).Infof("Enter allocate ...")
//	defer glog.V(4).Infof("Leaving allocate ...")
//
//	p := rus.Pop().(*policyapi.RequestUnit)
//	if p == nil {
//		glog.V(3).Infof("no pending units in Request <%v/%v>")
//		return false
//	}
//
//	assigned := false
//
//	for _, n := range ssn.Nodes {
//		if p.Resreq.Less(n.Idle) {
//			ssn.PublishDecision(&policyapi.Decision{
//				RequestID:     reqID,
//				RequestUnitID: p.ID,
//				Host:          n.Name,
//				Type:          policyapi.Allocate,
//			})
//			assigned = true
//
//			break
//		}
//	}
//
//	return assigned

//assigned := false
//for _, node := range nodes {
//	currentIdle := node.CurrentIdle()
//	if p.Request.LessEqual(currentIdle) {
//
//		// record the assignment temporarily in PodSet and Node
//		// this assignment will be accepted (min could be met in this time)
//		// or discarded (min could not be met in this time)
//		psi.assignPendingPod(p, node.Name)
//		node.AddUnAcceptedAllocated(p.Request)
//
//		assigned = true
//
//		unacceptedAssignedNodes[node.Name] = node
//
//		glog.V(3).Infof("assign <%v/%v> to <%s>: available <%v>, request <%v>",
//			p.Namespace, p.Name, p.NodeName, node.Idle, p.Request)
//		break
//	}
//}
//
//// the left resources can not meet any pod in this PodSet
//// (assume that all pods in same PodSet share same resource request)
//if !assigned {
//	// push pending pod back for consistent
//	psi.pushPendingPod(p)
//	break
//}
//
//min--
//
//
//if len(unacceptedAssignedNodes) == 0 {
//	// there is no nodes assigned pods this time
//	// the assignment is failed(no pod is assigned in this time)
//	return false
//}
//
//if min == 0 {
//	// min is met, accept all assignment this time
//	psi.acceptAssignedPods()
//	for _, node := range unacceptedAssignedNodes {
//		node.AcceptAllocated()
//	}
//	return true
//} else {
//	// min could not be met, discard all assignment this time
//	// to avoid PodSet get resources less than min
//	psi.discardAssignedPods()
//	for _, node := range unacceptedAssignedNodes {
//		node.DiscardAllocated()
//	}
//	return false
//}
//}
