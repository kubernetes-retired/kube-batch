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

package cache

import (
	"fmt"

	policyapi "github.com/kubernetes-incubator/kube-arbitrator/pkg/batchd/policy/api"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clientcache "k8s.io/client-go/tools/cache"
)

// podKey returns the string key of a pod.
func podKey(pod *v1.Pod) string {
	if key, err := clientcache.MetaNamespaceKeyFunc(pod); err != nil {
		return fmt.Sprintf("%v/%v", pod.Namespace, pod.Name)
	} else {
		return key
	}
}

func getPodOwner(pod *v1.Pod) types.UID {
	meta_pod, err := meta.Accessor(pod)
	if err != nil {
		return ""
	}

	controllerRef := metav1.GetControllerOf(meta_pod)
	if controllerRef != nil {
		return controllerRef.UID
	}

	return ""
}

func newResource(res *Resource) *policyapi.Resource {
	return &policyapi.Resource{
		MilliCPU: res.MilliCPU,
		Memory:   res.Memory,
		GPU:      res.GPU,
	}
}

func newRequestUnit(pi *PodInfo) *policyapi.RequestUnit {
	ru := &policyapi.RequestUnit{}

	ru.ID = fmt.Sprintf("%v/%v", pi.Namespace, pi.Name)
	ru.Resreq = newResource(pi.Request)

	return ru
}

func newRequest(psi *PodSet) *policyapi.Request {
	req := &policyapi.Request{}

	req.ID = fmt.Sprintf("%v/%v", psi.Namespace, psi.Name)

	for _, pi := range psi.Pending {
		ru := newRequestUnit(pi)
		ru.Status = policyapi.Pending
		req.Units[ru.ID] = ru
	}

	for _, pi := range psi.Assigned {
		ru := newRequestUnit(pi)
		ru.Status = policyapi.Bound
		req.Units[ru.ID] = ru
	}

	for _, pi := range psi.Running {
		ru := newRequestUnit(pi)
		ru.Status = policyapi.Running
		req.Units[ru.ID] = ru
	}

	return req
}

func newNode(ni *NodeInfo) *policyapi.Node {
	node := &policyapi.Node{}

	node.Name = ni.Name
	node.Allocatable = newResource(ni.Allocatable)
	for _, pi := range ni.Pods {
		ru := newRequestUnit(pi)
		node.Units[ru.ID] = ru
	}

	return node
}
