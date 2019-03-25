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

package validating

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/golang/glog"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kbv1 "github.com/kubernetes-sigs/kube-batch/pkg/apis/scheduling/v1alpha1"
	kbver "github.com/kubernetes-sigs/kube-batch/pkg/client/clientset/versioned"
	"github.com/kubernetes-sigs/kube-batch/pkg/client/clientset/versioned/typed/scheduling/v1alpha1"
	schedulerapi "github.com/kubernetes-sigs/kube-batch/pkg/scheduler/api"
	"github.com/kubernetes-sigs/kube-batch/pkg/webhook/api"
	"github.com/kubernetes-sigs/kube-batch/pkg/webhook/util"
)

var podGroupResource = metav1.GroupVersionResource{
	Group:    kbv1.GroupName,
	Version:  kbv1.GroupVersion,
	Resource: "podgroups",
}

// PodGroupValidator is used for validating whether we could a new PodGroup
type PodGroupValidator struct {
	KubeClient  *kbver.Clientset
	QueueClient v1alpha1.QueueInterface
	Locker      sync.Mutex
}

// NewPodGroupValidator returns a new PodGroupValidator for validating.
func NewPodGroupValidator(ctx api.AdmitterContext) (*PodGroupValidator, error) {
	return &PodGroupValidator{
		KubeClient:  ctx.KubeClient,
		QueueClient: ctx.KubeClient.SchedulingV1alpha1().Queues(),
	}, nil
}

// Admit is used to check whether we could create/delete the PodGroup, and update Queue's status.
func (pgv *PodGroupValidator) Admit(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	glog.V(2).Info("Admitting podgroup")
	if ar.Request.Resource != podGroupResource {
		err := fmt.Errorf("expect resource to be %s", podGroupResource)
		glog.Errorf("PodGroupValidator receives unexpected resource %+v", ar.Request)
		return util.ToAdmissionResponse(err)
	}

	raw := ar.Request.Object.Raw

	reviewResponse := &v1beta1.AdmissionResponse{}
	reviewResponse.Allowed = true

	podGroup, err := pgv.getPodGroup(ar)
	if err != nil {
		glog.Errorf("Failed to get PodGroup from %s: %v", raw, err)
		return util.ToAdmissionResponse(err)
	}

	if ar.Request.Operation == v1beta1.Create {
		if err := pgv.allocatePodGroup(podGroup); err != nil {
			glog.Errorf("Failed to create PodGroup %s/%s: %v", podGroup.Namespace, podGroup.Name, err)
			return util.ToAdmissionResponse(err)
		}
	} else if ar.Request.Operation == v1beta1.Delete {
		if err := pgv.deletePodGroup(podGroup); err != nil {
			glog.Errorf("Failed to delete PodGroup %s/%s: %v", podGroup.Namespace, podGroup.Name, err)
			return util.ToAdmissionResponse(err)
		}
	}

	return reviewResponse
}

// getPodGroup returns PodGroup according to admission request
func (pgv *PodGroupValidator) getPodGroup(ar *v1beta1.AdmissionReview) (*kbv1.PodGroup, error) {
	if ar.Request.Operation == v1beta1.Create {
		podGroup := kbv1.PodGroup{}
		if err := json.Unmarshal(ar.Request.Object.Raw, &podGroup); err != nil {
			glog.Errorf("Failed to unmarshal %s: %v", ar.Request.Object.Raw, err)
			return nil, err
		}
		return &podGroup, nil
	} else if ar.Request.Operation == v1beta1.Delete {
		return pgv.KubeClient.SchedulingV1alpha1().PodGroups(ar.Request.Namespace).Get(ar.Request.Name, metav1.GetOptions{})
	}
	return nil, fmt.Errorf("not a valid operations: %s", ar.Request.Operation)
}

// allocatePodGroup allocates podGroup in its specified queue, returns error if any.
func (pgv *PodGroupValidator) allocatePodGroup(podGroup *kbv1.PodGroup) error {
	pgv.Locker.Lock()
	defer pgv.Locker.Unlock()

	glog.V(4).Infof("Allocate pod group %s/%s in queue %s", podGroup.Namespace,
		podGroup.Name, podGroup.Spec.Queue)
	queue, err := pgv.getQueue(podGroup)
	if err != nil {
		return err
	}
	if queue == nil {
		return nil
	}

	if queue.Status.Allocated == nil {
		queue.Status.Allocated = make(corev1.ResourceList)
	}

	allocated, _ := queue.Status.Allocated[schedulerapi.PodGroup]
	capacity, exist := queue.Spec.Capability[schedulerapi.PodGroup]

	// If not set, default to infinite.
	if exist && capacity.Value() <= allocated.Value() {
		return fmt.Errorf("exceed podgroup limit(%d)", capacity.Value())
	}

	return pgv.updateQueue(queue, resource.NewQuantity(1, resource.DecimalSI))
}

// allocatePodGroup allocates podGroup in its specified queue, returns error if any.
func (pgv *PodGroupValidator) deletePodGroup(podGroup *kbv1.PodGroup) error {
	pgv.Locker.Lock()
	defer pgv.Locker.Unlock()

	glog.V(4).Infof("Delete pod group %s/%s in queue %s", podGroup.Namespace,
		podGroup.Name, podGroup.Spec.Queue)
	queue, err := pgv.getQueue(podGroup)
	if err != nil {
		return err
	}
	if queue == nil || queue.Status.Allocated == nil {
		return nil
	}

	return pgv.updateQueue(queue, resource.NewQuantity(-1, resource.DecimalSI))
}

// getQueue returns podGroup's queue
func (pgv *PodGroupValidator) getQueue(podGroup *kbv1.PodGroup) (*kbv1.Queue, error) {
	queueName := podGroup.Spec.Queue
	if len(queueName) == 0 {
		glog.Warningf("No queue specified for PodGroup %s/%s", podGroup.Namespace, podGroup.Name)
		return nil, nil
	}

	return pgv.QueueClient.Get(queueName, metav1.GetOptions{})
}

// updateQueue update queue's allocated PodGroup status according to the change
func (pgv *PodGroupValidator) updateQueue(queue *kbv1.Queue, change *resource.Quantity) error {
	allocated, _ := queue.Status.Allocated[schedulerapi.PodGroup]
	allocated.Add(*change)
	queue.Status.Allocated[schedulerapi.PodGroup] = allocated
	glog.V(4).Infof("Update queue %s status to %+v", queue.Name, queue.Status)
	_, err := pgv.QueueClient.Update(queue)
	return err
}
