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

package xqueuejob

import (
	"fmt"
	"github.com/golang/glog"
	"time"

	arbv1 "github.com/kubernetes-incubator/kube-arbitrator/pkg/apis/v1alpha1"
	"github.com/kubernetes-incubator/kube-arbitrator/pkg/client"
	"github.com/kubernetes-incubator/kube-arbitrator/pkg/client/clientset"
	arbinformers "github.com/kubernetes-incubator/kube-arbitrator/pkg/client/informers"
	informersv1 "github.com/kubernetes-incubator/kube-arbitrator/pkg/client/informers/v1"
	listersv1 "github.com/kubernetes-incubator/kube-arbitrator/pkg/client/listers/v1"
	"github.com/kubernetes-incubator/kube-arbitrator/pkg/controller/xqueuejob/queuejobresources"
	respod "github.com/kubernetes-incubator/kube-arbitrator/pkg/controller/xqueuejob/queuejobresources/pod"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const (
	// QueueJobNameLabel label string for queuejob name
	QueueJobNameLabel string = "xqueuejob-name"

	// ControllerUIDLabel label string for queuejob controller uid
	ControllerUIDLabel string = "controller-uid"
)

// controllerKind contains the schema.GroupVersionKind for this controller type.
var controllerKind = arbv1.SchemeGroupVersion.WithKind("XQueueJob")

//XController the XQueueJob Controller type
type XController struct {
	config           *rest.Config
	queueJobInformer informersv1.XQueueJobInformer

	// resources registered for the XQueueJob
	qjobRegisteredResources queuejobresources.RegisteredResources
	// controllers for these resources
	qjobResControls map[arbv1.ResourceType]queuejobresources.Interface

	clients    *kubernetes.Clientset
	arbclients *clientset.Clientset

	// A store of jobs
	queueJobLister listersv1.XQueueJobLister
	queueJobSynced func() bool

	// QueueJobs that need to be initialized
	// Add labels and selectors to XQueueJob
	initQueue *cache.FIFO

	// QueueJobs that need to sync up after initialization
	updateQueue *cache.FIFO

	// eventQueue that need to sync up
	eventQueue *cache.FIFO

	// Reference manager to manage membership of queuejob resource and its members
	refManager queuejobresources.RefManager
}

//RegisterAllQueueJobResourceTypes - gegisters all resources
func RegisterAllQueueJobResourceTypes(regs *queuejobresources.RegisteredResources) {
	respod.Register(regs)

}

func queueJobKey(obj interface{}) (string, error) {
	qj, ok := obj.(*arbv1.XQueueJob)
	if !ok {
		return "", fmt.Errorf("not a XQueueJob")
	}

	return fmt.Sprintf("%s/%s", qj.Namespace, qj.Name), nil
}

//NewXQueueJobController create new XQueueJob Controller
func NewXQueueJobController(config *rest.Config) *XController {
	cc := &XController{
		config:      config,
		clients:     kubernetes.NewForConfigOrDie(config),
		arbclients:  clientset.NewForConfigOrDie(config),
		eventQueue:  cache.NewFIFO(queueJobKey),
		initQueue:   cache.NewFIFO(queueJobKey),
		updateQueue: cache.NewFIFO(queueJobKey),
	}

	queueJobClient, _, err := client.NewClient(cc.config)
	if err != nil {
		panic(err)
	}

	cc.queueJobInformer = arbinformers.NewSharedInformerFactory(queueJobClient, 0).Batch().XQueueJobs()
	cc.qjobResControls = map[arbv1.ResourceType]queuejobresources.Interface{}
	RegisterAllQueueJobResourceTypes(&cc.qjobRegisteredResources)

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

	cc.queueJobInformer.Informer().AddEventHandler(
		cache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool {
				switch t := obj.(type) {
				case *arbv1.XQueueJob:
					glog.V(4).Infof("Filter XQueueJob name(%s) namespace(%s)\n", t.Name, t.Namespace)
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

	cc.queueJobSynced = cc.queueJobInformer.Informer().HasSynced

	//create sub-resource reference manager
	cc.refManager = queuejobresources.NewLabelRefManager()

	return cc
}

// Run start XQueueJob Controller
func (cc *XController) Run(stopCh chan struct{}) {
	// initialized
	createXQueueJobKind(cc.config)

	go cc.queueJobInformer.Informer().Run(stopCh)

	go cc.qjobResControls[arbv1.ResourceTypePod].Run(stopCh)

	cache.WaitForCacheSync(stopCh, cc.queueJobSynced)

	go wait.Until(cc.worker, time.Second, stopCh)

}

func (cc *XController) addQueueJob(obj interface{}) {
	qj, ok := obj.(*arbv1.XQueueJob)
	if !ok {
		glog.Errorf("obj is not XQueueJob")
		return
	}

	cc.enqueue(qj)
}

func (cc *XController) updateQueueJob(oldObj, newObj interface{}) {
	newQJ, ok := newObj.(*arbv1.XQueueJob)
	if !ok {
		glog.Errorf("newObj is not XQueueJob")
		return
	}

	cc.enqueue(newQJ)
}

func (cc *XController) deleteQueueJob(obj interface{}) {
	qj, ok := obj.(*arbv1.XQueueJob)
	if !ok {
		glog.Errorf("obj is not XQueueJob")
		return
	}

	cc.enqueue(qj)
}

func (cc *XController) enqueue(obj interface{}) {
	err := cc.eventQueue.Add(obj)
	if err != nil {
		glog.Errorf("Fail to enqueue XQueueJob to updateQueue, err %#v", err)
	}
}

func (cc *XController) worker() {
	if _, err := cc.eventQueue.Pop(func(obj interface{}) error {
		var queuejob *arbv1.XQueueJob
		switch v := obj.(type) {
		case *arbv1.XQueueJob:
			queuejob = v
		default:
			glog.Errorf("Un-supported type of %v", obj)
			return nil
		}

		if queuejob == nil {
			if acc, err := meta.Accessor(obj); err != nil {
				glog.Warningf("Failed to get XQueueJob for %v/%v", acc.GetNamespace(), acc.GetName())
			}

			return nil
		}

		// sync XQueueJob
		if err := cc.syncQueueJob(queuejob); err != nil {

			glog.Errorf("Failed to sync XQueueJob %s %s, err %#v", queuejob.Name, queuejob.Namespace, err)

			// If any error, requeue it.
			return err
		}

		return nil
	}); err != nil {
		glog.Errorf("Fail to pop item from updateQueue, err %#v", err)
		return
	}
}

func (cc *XController) syncQueueJob(qj *arbv1.XQueueJob) error {
	glog.Infof("Sync QJ %s %s", qj.Name, qj.Namespace)
	queueJob, err := cc.queueJobLister.XQueueJobs(qj.Namespace).Get(qj.Name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			glog.V(3).Infof("Job has been deleted: %v", qj.Name)
			return nil
		}
		return err
	}

	return cc.manageQueueJob(queueJob)
}

func (cc *XController) manageQueueJob(qj *arbv1.XQueueJob) error {
	var err error
	startTime := time.Now()
	defer func() {
		glog.Infof("Finished syncing queue job %q (%v)", qj.Name, time.Now().Sub(startTime))
	}()

	if qj.DeletionTimestamp != nil {
		// cleanup resources for running job
		glog.Infof("Cleaning up resources for the QueueJob................ %s", qj.Name)
		for _, ar := range qj.Spec.AggrResources.Items {
			ar.Replicas = 0
		}
		qj.Status.MinAvailable = -1
		if _, err := cc.arbclients.ArbV1().XQueueJobs(qj.Namespace).Update(qj); err != nil {
			glog.Errorf("Failed to update status of XQueueJob %v/%v: %v",
				qj.Namespace, qj.Name, err)
			return err
		}

		err = cc.Cleanup(qj)
		if err != nil {
			return err
		}
		return nil
	}

	// we call sync for each controller
	for _, ar := range qj.Spec.AggrResources.Items {
		cc.qjobResControls[ar.Type].EnqueueSyncQueueJob(qj, &ar)
	}

	// TODO(k82cn): replaced it with `UpdateStatus`
	if _, err := cc.arbclients.ArbV1().XQueueJobs(qj.Namespace).Update(qj); err != nil {
		glog.Errorf("Failed to update status of XQueueJob %v/%v: %v",
			qj.Namespace, qj.Name, err)
		return err
	}

	return err
}

//Cleanup function
func (cc *XController) Cleanup(queuejob *arbv1.XQueueJob) error {
	if queuejob.Spec.AggrResources.Items != nil {
		// we call clean-up for each controller
		for _, ar := range queuejob.Spec.AggrResources.Items {
			cc.qjobResControls[ar.Type].Cleanup(queuejob, &ar)
		}
	}

	return nil
}
