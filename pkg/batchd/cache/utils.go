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

func getPodOwner(pod *v1.Pod, groupLabel string) types.UID {
	// The pod with different owner reference may belong to the same PodSet,
	//
	// K8S Deployment
	// A deployment may contain more than one RS when rolling-upgrade happens,
	// pods in different RS have different owner references, so kube-batchd will
	// group them into different PodSet. However, these pods belong to the same
	// deployment and should be in one PodSet.
	//
	// Kubeflow job
	// In a tfjob, kubeflow create Master/PS/Worker as a k8s job and each job
	// contains one pod. kube-batchd will group these pods into different PodSet
	// because of different owner references. However, they belong to the same
	// tfjob and should be in one PodSet.
	//
	// It is hard to get a k8s resources by its uid currently, use label
	// groupLabel as its owner reference temporarily to group pods into PodSet
	// it is user's responsibility to make the label unique
	// TODO(jinzhej): get the root owner references for a pod, not use label
	if v, ok := pod.Labels[groupLabel]; ok {
		return types.UID(v)
	}

	// get pod owner reference directly if label groupLabel is empty
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
