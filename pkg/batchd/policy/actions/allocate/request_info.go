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

//
//type requestInfo struct {
//	request *policyapi.Request
//
//	pending *util.PriorityQueue
//}
//
//func newRequestInfo(req *policyapi.Request) *requestInfo {
//	reqInfo := &requestInfo{
//		request: req,
//		pending: util.NewPriorityQueue(nil),
//	}
//
//	for _, ru := range req.Units {
//		if ru.Status == policyapi.Pending {
//			reqInfo.pending.Push(ru)
//		}
//	}
//
//	return reqInfo
//}
//
//func (ri *requestInfo) nextPending() *policyapi.RequestUnit {
//	return ri.pending.Pop().(*policyapi.RequestUnit)
//}

//
//func (psi *podSetInfo) assignPendingPod(p *cache.PodInfo, nodeName string) {
//	// assign node to pending pod temporarily
//	psi.unacceptedAllocated.Add(p.Request)
//	p.NodeName = nodeName
//	psi.unacceptedAssignedPods = append(psi.unacceptedAssignedPods, p)
//
//	glog.V(3).Infof("PodSet <%v/%v> after assignment: priority <%f>, dominant resource <%v>",
//		psi.podSet.Namespace, psi.podSet.Name, psi.share, psi.dominantResource)
//}
//
//func (psi *podSetInfo) popPendingPod() *cache.PodInfo {
//	if psi.pendingSorted.Empty() {
//		return nil
//	}
//
//	pi := psi.pendingSorted.Pop().(*cache.PodInfo)
//
//	return pi
//}
//
//func (psi *podSetInfo) pushPendingPod(p *cache.PodInfo) {
//	psi.pendingSorted.Push(p, -float64(p.Priority))
//}
//
//func (psi *podSetInfo) insufficientMinAvailable() int {
//	insufficient := 0
//	if len(psi.podSet.Running)+len(psi.podSet.Assigned) < psi.podSet.MinAvailable {
//		insufficient = psi.podSet.MinAvailable - len(psi.podSet.Running) - len(psi.podSet.Assigned)
//	}
//	return insufficient
//}
//
//func (psi *podSetInfo) acceptAssignedPods() {
//	if len(psi.unacceptedAssignedPods) == 0 {
//		return
//	}
//
//	// accept temporary assigned Pods
//	// put them to PodSet assigned queue
//	psi.podSet.Assigned = append(psi.podSet.Assigned, psi.unacceptedAssignedPods...)
//	psi.unacceptedAssignedPods = make([]*cache.PodInfo, 0)
//
//	// update allocate resource for consistent
//	psi.allocated.Add(psi.unacceptedAllocated)
//	psi.unacceptedAllocated = cache.EmptyResource()
//
//	// update podset share
//	psi.share = psi.calculateShare(psi.dominantResource)
//}
//
//func (psi *podSetInfo) discardAssignedPods() {
//	if len(psi.unacceptedAssignedPods) == 0 {
//		return
//	}
//
//	// clean assigned node
//	for _, p := range psi.unacceptedAssignedPods {
//		p.NodeName = ""
//	}
//
//	// discard temporary assigned Pods
//	// put them back to PodSet pending queue
//	for _, p := range psi.unacceptedAssignedPods {
//		psi.pendingSorted.Push(p, -float64(p.Priority))
//	}
//	psi.unacceptedAssignedPods = make([]*cache.PodInfo, 0)
//
//	psi.unacceptedAllocated = cache.EmptyResource()
//}
//
//func (psi *podSetInfo) calculateShare(rn v1.ResourceName) float64 {
//	return psi.allocated.Get(rn) / psi.total.Get(rn)
//}
