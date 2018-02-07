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

package drf

import (
	"sort"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/kube-arbitrator/pkg/batchd/cache"
	"github.com/kubernetes-incubator/kube-arbitrator/pkg/batchd/policy/util"
)

// PolicyName is the name of drf policy; it'll be use for any case
// that need a name, e.g. default policy, register drf policy.
var PolicyName = "drf"

type drfScheduler struct {
}

func New() *drfScheduler {
	return &drfScheduler{}
}

func (drf *drfScheduler) Name() string {
	return PolicyName
}

func (drf *drfScheduler) Initialize() {}

func (drf *drfScheduler) Allocate(queues []*cache.QueueInfo, nodes []*cache.NodeInfo) []*cache.QueueInfo {
	glog.V(4).Infof("Enter Allocate ...")
	defer glog.V(4).Infof("Leaving Allocate ...")

	total := cache.EmptyResource()
	for _, n := range nodes {
		total.Add(n.Allocatable)
	}

	dq := util.NewDictionaryQueue()
	for _, c := range queues {
		for _, ps := range c.PodSets {
			psi := newPodSetInfo(ps, total)
			dq.Push(util.NewDictionaryItem(psi, psi.podSet.Name))
		}
	}

	// assign MinAvailable of each podSet first by chronologically
	sort.Sort(dq)
	meetAllMinAvailable := true
	for _, q := range dq {
		psi := q.Value.(*podSetInfo)
		assignedPods := make(map[string]string)
		for {
			if psi.meetMinAvailable() {
				glog.V(3).Infof("MinAvailable of podset %s/%s is met, assign next podset",
					psi.podSet.Namespace, psi.podSet.Name)
				break
			}

			p := psi.popPendingPod()
			if p == nil {
				glog.V(3).Infof("no pending Pod in PodSet <%v/%v>",
					psi.podSet.Namespace, psi.podSet.Name)
				break
			}

			assigned := false
			for _, node := range nodes {
				if p.Request.LessEqual(node.Idle) {
					psi.assignPendingPod(p, node.Name)
					node.Idle.Sub(p.Request)

					assigned = true

					assignedPods[p.Name] = node.Name

					glog.V(3).Infof("assign <%v/%v> to <%s>: available <%v>, request <%v>",
						p.Namespace, p.Name, p.NodeName, node.Idle, p.Request)
					break
				}
			}

			// the left resources can not meet any pod in this podset
			if !assigned {
				break
			}
		}

		// If policy could not assign minAvailable for a PodSet, then recycle assigned resources in
		// this loop even if the PodSet was assigned resources before it is to handle a case that PodSet
		// meet minAvailable in the past and then become under minAvailable due to some pods finish
		if !psi.meetMinAvailable() {
			for podName, nodeName := range assignedPods {
				pi := psi.resetAndGetAssignedPod(podName)
				if pi == nil {
					continue
				}
				for _, n := range nodes {
					if nodeName != n.Name {
						continue
					}
					n.Idle.Add(pi.Request)
				}
			}
			meetAllMinAvailable = false
		}
	}

	// If minAvailable of all PodSets are meet, then assign left resources to these PodSets by DRF
	// Otherwise, policy only allocate minAvailable for each PodSet. The left resources will not be
	// assigned to the left PodSet.
	if !meetAllMinAvailable {
		return queues
	}

	// build priority queue after assign minAvailable
	pq := util.NewPriorityQueue()
	for _, q := range dq {
		psi := q.Value.(*podSetInfo)
		pq.Push(psi, psi.share)
	}

	for {
		if pq.Empty() {
			break
		}

		psi := pq.Pop().(*podSetInfo)

		glog.V(3).Infof("try to allocate resources to PodSet <%v/%v>",
			psi.podSet.Namespace, psi.podSet.Name)

		p := psi.popPendingPod()

		// If no pending pods, skip this PodSet without push it back.
		if p == nil {
			glog.V(3).Infof("no pending Pod in PodSet <%v/%v>",
				psi.podSet.Namespace, psi.podSet.Name)
			continue
		}

		assigned := false
		for _, node := range nodes {
			if p.Request.LessEqual(node.Idle) {
				psi.assignPendingPod(p, node.Name)
				node.Idle.Sub(p.Request)

				assigned = true

				glog.V(3).Infof("assign <%v/%v> to <%s>: available <%v>, request <%v>",
					p.Namespace, p.Name, p.NodeName, node.Idle, p.Request)
				break
			}
		}

		if assigned {
			pq.Push(psi, psi.share)
		} else {
			// push pending pod back for consistent
			psi.pushPendingPod(p)
			// If no assignment, did not push PodSet back as no node can be used.
			glog.V(3).Infof("no node was assigned to <%v/%v> with request <%v>",
				p.Namespace, p.Name, p.Request)
		}
	}

	return queues
}

func (drf *drfScheduler) UnInitialize() {}
