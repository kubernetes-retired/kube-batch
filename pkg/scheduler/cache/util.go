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

package cache

import (
	"strconv"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/glog"

	"github.com/kubernetes-sigs/kube-batch/pkg/apis/scheduling/v1alpha1"
	"github.com/kubernetes-sigs/kube-batch/pkg/apis/scheduling/v1alpha2"
	"github.com/kubernetes-sigs/kube-batch/pkg/apis/utils"
	"github.com/kubernetes-sigs/kube-batch/pkg/scheduler/api"
)

const (
	shadowPodGroupKey = "volcano/shadow-pod-group"
)

func shadowPodGroup(pg *api.PodGroup) bool {
	if pg == nil {
		return true
	}

	_, found := pg.Annotations[shadowPodGroupKey]

	return found
}

func createShadowPodGroup(pod *v1.Pod) *api.PodGroup {
	jobID := api.JobID(utils.GetController(pod))
	if len(jobID) == 0 {
		jobID = api.JobID(pod.UID)
	}

	// Deriving min member for the shadow pod group from pod annotations.
	//
	// Annotation from the newer API has priority over annotation from the old API.
	//
	// By default, if no annotation is provided, min member is 1.
	minMember := 1
	if annotationValue, found := pod.Annotations[v1alpha1.GroupMinMemberAnnotationKey]; found {
		if integerValue, err := strconv.Atoi(annotationValue); err == nil {
			minMember = integerValue
		} else {
			glog.Errorf("Pod %s/%s has illegal value %q for annotation %q",
				pod.Namespace, pod.Name, annotationValue, v1alpha1.GroupMinMemberAnnotationKey)
		}
	}
	if annotationValue, found := pod.Annotations[v1alpha2.GroupMinMemberAnnotationKey]; found {
		if integerValue, err := strconv.Atoi(annotationValue); err == nil {
			minMember = integerValue
		} else {
			glog.Errorf("Pod %s/%s has illegal value %q for annotation %q",
				pod.Namespace, pod.Name, annotationValue, v1alpha2.GroupMinMemberAnnotationKey)
		}
	}

	return &api.PodGroup{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: pod.Namespace,
			Name:      string(jobID),
			Annotations: map[string]string{
				shadowPodGroupKey: string(jobID),
			},
			CreationTimestamp: pod.CreationTimestamp,
		},
		Spec: api.PodGroupSpec{
			MinMember:         int32(minMember),
			PriorityClassName: pod.Spec.PriorityClassName,
		},
	}
}
