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

package queuejob

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"

	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"


	"github.com/kubernetes-incubator/kube-arbitrator/pkg/queuejob-ctrl/controller/queuejobresources"
	"github.com/kubernetes-incubator/kube-arbitrator/pkg/queuejob-ctrl/controller/queuejobresources"
	respod "github.com/kubernetes-incubator/kube-arbitrator/pkg/queuejob-ctrl/controller/queuejobresources/pod"

	"github.com/kubernetes-incubator/kube-arbitrator/pkg/apis/utils"
	arbv1 "github.com/kubernetes-incubator/kube-arbitrator/pkg/apis/v1alpha1"
	"github.com/kubernetes-incubator/kube-arbitrator/pkg/client"
	"github.com/kubernetes-incubator/kube-arbitrator/pkg/client/clientset"
	arbinformers "github.com/kubernetes-incubator/kube-arbitrator/pkg/client/informers"
	informersv1 "github.com/kubernetes-incubator/kube-arbitrator/pkg/client/informers/v1"
	listersv1 "github.com/kubernetes-incubator/kube-arbitrator/pkg/client/listers/v1"
)

const (
	// QueueJobNameLabel label string for queuejob name
	QueueJobNameLabel string = "queuejob-name"

	// ControllerUIDLabel label string for queuejob controller uid
	ControllerUIDLabel string = "controller-uid"
)

// controllerKind contains the schema.GroupVersionKind for this controller type.
var controllerKind = arbv1.SchemeGroupVersion.WithKind("QueueJob")

// Controller the QueueJob Controller type
type Controller struct {
	config           *rest.Config
	queueJobInformer informersv1.QueueJobInformer
	// resources registered for the QueueJob
	qjobRegisteredResources queuejobresources.RegisteredResources
	// controllers for these resources
	qjobResControls         map[arbv1.ResourceType]queuejobresources.Interface
	
	clients          *kubernetes.Clientset
	arbclients       *clientset.Clientset

	// A store of jobs
	queueJobLister listersv1.QueueJobLister
	queueJobSynced func() bool

	// QueueJobs that need to be initialized
	// Add labels and selectors to QueueJob
	initQueue *cache.FIFO

	// QueueJobs that need to sync up after initialization
	updateQueue *cache.FIFO

	// A map that store PDB name for each QueueJob
	queueJobToPDB map[string]string
	// Reference manager to manage membership of queuejob resource and its members
	refManager queuejobresources.RefManager
}

func RegisterAllQueueJobResourceTypes(regs *queuejobresources.RegisteredResources) {
	respod.Register(regs)

}

// NewController create new QueueJob Controller
func NewController(config *rest.Config) *Controller {
	cc := &Controller{
		config:             config,
		clients:            kubernetes.NewForConfigOrDie(config),
		arbclients:         clientset.NewForConfigOrDie(config),
		initQueue:          cache.NewFIFO(queueJobKey),
		updateQueue:        cache.NewFIFO(queueJobKey),
		queueJobToPDB:      make(map[string]string),
	}

	queueJobClient, _, err := client.NewClient(cc.config)
	if err != nil {
		panic(err)
	}
	
	//initialize pod sub-resource control
	resControlPod, found, err := cc.qjobRegisteredResources.InitQueueJobResource(arbv1.ResourceTypePod, config)
	if err != nil {
		glog.Errorf("fail to create queuejob resource control")
		return nil
	}
	if !found {
		glog.Errorf("queuejob resource type Pod not found")
		return nil
	}
	cc.qjobResControls[arbv1.ResourceTypePod] = resControlPod
	
	//register and initialize sub-resources controls
	RegisterAllQueueJobResourceTypes(&arbv1.qjobRegisteredResources)
	cc.qjobResControls = map[arbv1.ResourceType]queuejobresources.Interface{}

	cc.queueJobInformer = arbinformers.NewSharedInformerFactory(queueJobClient, 0).QueueJob().QueueJobs()
	cc.queueJobInformer.Informer().AddEventHandler(
		cache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool {
				switch t := obj.(type) {
				case *arbv1.QueueJob:
					glog.V(4).Infof("Filter QueueJob name(%s) namespace(%s)\n", t.Name, t.Namespace)
					return true
				default:
					return false
				}
			},
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc:    cc.addQueueJob,
				UpdateFunc: cc.updateQueueJob,
				DeleteFunc: cc.deleteQueueJob,
			},
	})
	cc.queueJobLister = cc.queueJobInformer.Lister()
	
	//create sub-resource reference manager
	cc.refManager = queuejobresources.NewLabelRefManager()

	return cc
}

// Run start QueueJob Controller
func (cc *Controller) Run(stopCh chan struct{}) {
	// initialized
	createQueueJobCRD(cc.config)

	go cc.queueJobInformer.Informer().Run(stopCh)
	
	go cc.qjobResControls[arbv1.ResourceTypePod].Run(stopCh)

	go wait.Until(cc.initWorker, time.Second, stopCh)
	go wait.Until(cc.updateWorker, time.Second, stopCh)
}

func (cc *Controller) addQueueJob(obj interface{}) {
	qj, ok := obj.(*arbv1.QueueJob)
	if !ok {
		glog.Errorf("obj is not QueueJob")
		return
	}

	cc.enqueueInitQueue(qj)
}

func (cc *Controller) updateQueueJob(oldObj, newObj interface{}) {
	oldQJ, ok := oldObj.(*arbv1.QueueJob)
	if !ok {
		glog.Errorf("oldObj is not QueueJob")
		return
	}
	newQJ, ok := newObj.(*arbv1.QueueJob)
	if !ok {
		glog.Errorf("newObj is not QueueJob")
		return
	}

	cc.enqueue(newQJ)
}

func (cc *Controller) deleteQueueJob(obj interface{}) {
	qj, ok := obj.(*arbv1.QueueJob)
	if !ok {
		glog.Errorf("obj is not QueueJob")
		return
	}

	cc.enqueue(qj)
}

func (cc *Controller) enqueue(obj interface{}) {
	err := cc.eventQueue.Add(obj)
	if err != nil {
		glog.Errorf("Fail to enqueue QueueJob to updateQueue, err %#v", err)
	}
}

func (cc *Controller) worker() {
	if _, err := cc.eventQueue.Pop(func(obj interface{}) error {
		var queuejob *arbv1.QueueJob
		switch v := obj.(type) {
		case *arbv1.QueueJob:
			queuejob = v
		default:
			glog.Errorf("Un-supported type of %v", obj)
			return nil
		}

		if queuejob == nil {
			if acc, err := meta.Accessor(obj); err != nil {
				glog.Warningf("Failed to get QueueJob for %v/%v", acc.GetNamespace(), acc.GetName())
			}

			return nil
		}

		// sync QueueJob
		if err := cc.syncQueueJob(queuejob); err != nil {
			glog.Errorf("Failed to sync QueueJob %s, err %#v", queuejob.Name, err)
			// If any error, requeue it.
			return err
		}

		return nil
	}); err != nil {
		glog.Errorf("Fail to pop item from updateQueue, err %#v", err)
		return
	}
}

func (cc *Controller) syncQueueJob(qj *arbv1.QueueJob) error {
	queueJob, err := cc.queueJobLister.QueueJobs(qj.Namespace).Get(qj.Name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			glog.V(3).Infof("Job has been deleted: %v", qj.Name)
			return nil
		}
		return err
	}

	return cc.manageQueueJob(queueJob, pods)
}

// manageQueueJob is the core method responsible for managing the number of running
// pods according to what is specified in the job.Spec.
// Does NOT modify <activePods>.
func (cc *Controller) manageQueueJob(qj *arbv1.QueueJob, pods []*v1.Pod) error {
	
	startTime := time.Now()
	defer func() {
		glog.V(4).Infof("Finished syncing queue job %q (%v)", key, time.Now().Sub(startTime))
	}()

	if qj.DeletionTimestamp != nil {
		// cleanup resources for running job
		err = qjm.Cleanup(qj)
		if err != nil {
			return err
		}
		//empty finalizers and delete the queuejob again
		accessor, err := meta.Accessor(qj)
		if err != nil {
			return err
		}
		accessor.SetFinalizers(nil)

		var result arbv1.QueueJob
		return qjm.qjobClient.Put().
			Namespace(ns).Resource(arbv1.QueueJobPlural).
			Name(name).Body(qj).Do().Into(&result)
	}
	
	// we call sync for each controller
	for _, ar := range queuejob.Spec.AggrResources.Items {
			cc.qjobResControls[ar.Type].Sync(qj)
	}
	
	// TODO(k82cn): replaced it with `UpdateStatus`
	if _, err := cc.arbclients.ArbV1().QueueJobs(qj.Namespace).Update(qj); err != nil {
		glog.Errorf("Failed to update status of QueueJob %v/%v: %v",
			qj.Namespace, qj.Name, err)
		return err
	}

	return err
}

func (cc *Controller) Cleanup(queuejob *arbv1.QueueJob) error {
	if queuejob.Spec.AggrResources.Items != nil {
		// we call clean-up for each controller
		for _, ar := range queuejob.Spec.AggrResources.Items {
			cc.qjobResControls[ar.Type].Cleanup(queuejob, &ar)
		}
	}
	return nil
}
