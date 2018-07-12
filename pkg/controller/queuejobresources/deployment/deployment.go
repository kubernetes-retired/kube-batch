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

package deployment

import (
	"fmt"
	"github.com/golang/glog"
	arbv1 "github.com/kubernetes-incubator/kube-arbitrator/pkg/apis/v1alpha1"
	"github.com/kubernetes-incubator/kube-arbitrator/pkg/client/clientset"
	"github.com/kubernetes-incubator/kube-arbitrator/pkg/controller/queuejobresources"
	schedulerapi "github.com/kubernetes-incubator/kube-arbitrator/pkg/scheduler/api"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	extinformer "k8s.io/client-go/informers/apps/v1beta1"
	"k8s.io/client-go/kubernetes"
	extlister "k8s.io/client-go/listers/apps/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"sync"
	"time"
)

var queueJobKind = arbv1.SchemeGroupVersion.WithKind("XQueueJob")
var queueJobName = "xqueuejob.arbitrator.k8s.io"

const (
	// QueueJobNameLabel label string for queuejob name
	QueueJobNameLabel string = "xqueuejob-name"

	// ControllerUIDLabel label string for queuejob controller uid
	ControllerUIDLabel string = "controller-uid"
)

//QueueJobResDeployment contains the resources of this queuejob
type QueueJobResDeployment struct {
	clients    *kubernetes.Clientset
	arbclients *clientset.Clientset
	// A store of services, populated by the serviceController
	serviceStore   extlister.DeploymentLister
	deployInformer extinformer.DeploymentInformer
	rtScheme       *runtime.Scheme
	jsonSerializer *json.Serializer
	// Reference manager to manage membership of queuejob resource and its members
	refManager queuejobresources.RefManager
}

//Register registers a queue job resource type
func Register(regs *queuejobresources.RegisteredResources) {
	regs.Register(arbv1.ResourceTypeDeployment, func(config *rest.Config) queuejobresources.Interface {
		return NewQueueJobResDeployment(config)
	})
}

//NewQueueJobResDeployment returns a new deployment controller
func NewQueueJobResDeployment(config *rest.Config) queuejobresources.Interface {
	qjrd := &QueueJobResDeployment{
		clients:    kubernetes.NewForConfigOrDie(config),
		arbclients: clientset.NewForConfigOrDie(config),
	}

	qjrd.deployInformer = informers.NewSharedInformerFactory(qjrd.clients, 0).Apps().V1beta1().Deployments()
	qjrd.deployInformer.Informer().AddEventHandler(
		cache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool {
				switch obj.(type) {
				case *v1beta1.Deployment:
					return true
				default:
					return false
				}
			},
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc:    qjrd.addDeployment,
				UpdateFunc: qjrd.updateDeployment,
				DeleteFunc: qjrd.deleteDeployment,
			},
		})

	qjrd.rtScheme = runtime.NewScheme()
	v1.AddToScheme(qjrd.rtScheme)
	v1beta1.AddToScheme(qjrd.rtScheme)
	qjrd.jsonSerializer = json.NewYAMLSerializer(json.DefaultMetaFactory, qjrd.rtScheme, qjrd.rtScheme)

	qjrd.refManager = queuejobresources.NewLabelRefManager()

	return qjrd
}

func (qjrPod *QueueJobResDeployment) GetAggregatedResources(job *arbv1.XQueueJob) *schedulerapi.Resource {
        return schedulerapi.EmptyResource()
}

//Run the main goroutine responsible for watching and services.
func (qjrService *QueueJobResDeployment) Run(stopCh <-chan struct{}) {

	qjrService.deployInformer.Informer().Run(stopCh)
}

func (qjrService *QueueJobResDeployment) addDeployment(obj interface{}) {

	return
}

func (qjrService *QueueJobResDeployment) updateDeployment(old, cur interface{}) {

	return
}

func (qjrService *QueueJobResDeployment) deleteDeployment(obj interface{}) {

	return
}


// Parse queue job api object to get Service template
func (qjrService *QueueJobResDeployment) getDeploymentTemplate(qjobRes *arbv1.XQueueJobResource) (*v1beta1.Deployment, error) {
	serviceGVK := schema.GroupVersion{Group: v1beta1.GroupName, Version: "v1beta1"}.WithKind("Deployment")
	obj, _, err := qjrService.jsonSerializer.Decode(qjobRes.Template.Raw, &serviceGVK, nil)
	if err != nil {
		return nil, err
	}

	service, ok := obj.(*v1beta1.Deployment)
	if !ok {
		return nil, fmt.Errorf("Queuejob resource not defined as a Service")
	}

	return service, nil

}

func (qjrService *QueueJobResDeployment) createDeploymentWithControllerRef(namespace string, service *v1beta1.Deployment, controllerRef *metav1.OwnerReference) error {
	glog.V(4).Infof("==========create service: %s,  %+v \n", namespace, service)
	if controllerRef != nil {
		service.OwnerReferences = append(service.OwnerReferences, *controllerRef)
	}

	if _, err := qjrService.clients.ExtensionsV1beta1().Deployments(namespace).Create(service); err != nil {
		return err
	}

	return nil
}

func (qjrService *QueueJobResDeployment) delDeployment(namespace string, name string) error {

	glog.V(4).Infof("==========delete service: %s,  %s \n", namespace, name)
	if err := qjrService.clients.ExtensionsV1beta1().Deployments(namespace).Delete(name, nil); err != nil {
		return err
	}

	return nil
}

//SyncQueueJob syncs the resources of this queuejob
func (qjrService *QueueJobResDeployment) SyncQueueJob(queuejob *arbv1.XQueueJob, qjobRes *arbv1.XQueueJobResource) error {

	startTime := time.Now()
	defer func() {
		glog.V(4).Infof("Finished syncing queue job resource %q (%v)", qjobRes.Template, time.Now().Sub(startTime))
	}()

	services, err := qjrService.getDeploymentsForQueueJobRes(qjobRes, queuejob)
	if err != nil {
		return err
	}

	serviceLen := len(services)
	replicas := qjobRes.Replicas

	diff := int(replicas) - int(serviceLen)

	glog.V(4).Infof("QJob: %s had %d services and %d desired services", queuejob.Name, replicas, serviceLen)

	if diff > 0 {
		template, err := qjrService.getDeploymentTemplate(qjobRes)
		if err != nil {
			glog.Errorf("Cannot read template from resource %+v %+v", qjobRes, err)
			return err
		}
		//TODO: need set reference after Service has been really added
		tmpService := v1.Service{}
		err = qjrService.refManager.AddReference(qjobRes, &tmpService)
		if err != nil {
			glog.Errorf("Cannot add reference to service resource %+v", err)
			return err
		}

		if template.Labels == nil {
			template.Labels = map[string]string{}
		}
		for k, v := range tmpService.Labels {
			template.Labels[k] = v
		}
		wait := sync.WaitGroup{}
		wait.Add(int(diff))
		for i := 0; i < diff; i++ {
			go func() {
				defer wait.Done()
				err := qjrService.createDeploymentWithControllerRef(queuejob.Namespace, template, metav1.NewControllerRef(queuejob, queueJobKind))
				if err != nil && errors.IsTimeout(err) {
					return
				}
				if err != nil {
					defer utilruntime.HandleError(err)
				}
			}()
		}
		wait.Wait()
	}

	return nil
}

func (qjrService *QueueJobResDeployment) getDeploymentsForQueueJob(j *arbv1.XQueueJob) ([]*v1beta1.Deployment, error) {
	servicelist, err := qjrService.clients.ExtensionsV1beta1().Deployments(j.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	services := []*v1beta1.Deployment{}
	for i, service := range servicelist.Items {
		metaService, err := meta.Accessor(&service)
		if err != nil {
			return nil, err
		}

		controllerRef := metav1.GetControllerOf(metaService)
		if controllerRef != nil {
			if controllerRef.UID == j.UID {
				services = append(services, &servicelist.Items[i])
			}
		}
	}
	return services, nil

}

func (qjrService *QueueJobResDeployment) getDeploymentsForQueueJobRes(qjobRes *arbv1.XQueueJobResource, j *arbv1.XQueueJob) ([]*v1beta1.Deployment, error) {

	services, err := qjrService.getDeploymentsForQueueJob(j)
	if err != nil {
		return nil, err
	}

	myServices := []*v1beta1.Deployment{}
	for i, service := range services {
		if qjrService.refManager.BelongTo(qjobRes, service) {
			myServices = append(myServices, services[i])
		}
	}

	return myServices, nil

}

func (qjrService *QueueJobResDeployment) deleteQueueJobResDeployments(qjobRes *arbv1.XQueueJobResource, queuejob *arbv1.XQueueJob) error {
	job := *queuejob

	activeServices, err := qjrService.getDeploymentsForQueueJobRes(qjobRes, queuejob)
	if err != nil {
		return err
	}

	active := int32(len(activeServices))

	wait := sync.WaitGroup{}
	wait.Add(int(active))
	for i := int32(0); i < active; i++ {
		go func(ix int32) {
			defer wait.Done()
			if err := qjrService.delDeployment(queuejob.Namespace, activeServices[ix].Name); err != nil {
				defer utilruntime.HandleError(err)
				glog.V(2).Infof("Failed to delete %v, queue job %q/%q deadline exceeded", activeServices[ix].Name, job.Namespace, job.Name)
			}
		}(i)
	}
	wait.Wait()

	return nil
}

//Cleanup deletes all resources with this contorller
func (qjrService *QueueJobResDeployment) Cleanup(queuejob *arbv1.XQueueJob, qjobRes *arbv1.XQueueJobResource) error {
	return qjrService.deleteQueueJobResDeployments(qjobRes, queuejob)
}
