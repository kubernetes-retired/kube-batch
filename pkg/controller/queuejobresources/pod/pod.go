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

package pod

import (
	"fmt"
	"sync"
	"time"
	uuid "github.com/satori/uuid"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	arbv1 "github.com/kubernetes-incubator/kube-arbitrator/pkg/apis/v1alpha1"
	"github.com/kubernetes-incubator/kube-arbitrator/pkg/controller/maputils"
	"github.com/kubernetes-incubator/kube-arbitrator/pkg/controller/queuejobresources"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	corev1informer "k8s.io/client-go/informers/core/v1"
	"github.com/kubernetes-incubator/kube-arbitrator/pkg/client/clientset"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

var queueJobKind = arbv1.SchemeGroupVersion.WithKind("XQueueJob")
var queueJobName = "xqueuejob.arbitrator.k8s.io"

const (
	// QueueJobNameLabel label string for queuejob name
	QueueJobNameLabel string = "xqueuejob-name"

	// ControllerUIDLabel label string for queuejob controller uid
	ControllerUIDLabel string = "controller-uid"
)

type QueueJobResPod struct {

	clients          *kubernetes.Clientset
	arbclients       *clientset.Clientset

	// A TTLCache of pod creates/deletes each rc expects to see

	// A store of pods, populated by the podController
	podStore       corelisters.PodLister
	podInformer    corev1informer.PodInformer
	
	podSynced func() bool

	// A counter that stores the current terminating pod no of QueueJob
	// this is used to avoid to re-create the pods of a QueueJob before
	// all the old resources are terminated
	deletedResourcesCounter *maputils.SyncCounterMap
	rtScheme       *runtime.Scheme
	jsonSerializer *json.Serializer

	// Reference manager to manage membership of queuejob resource and its members
	refManager queuejobresources.RefManager
	// A counter that store the current terminating pods no of QueueJob
	// this is used to avoid to re-create the pods of a QueueJob before
	// all the old pods are terminated
	deletedPodsCounter *maputils.SyncCounterMap
}

// Register registers a queue job resource type
func Register(regs *queuejobresources.RegisteredResources) {
	regs.Register(arbv1.ResourceTypePod, func(config *rest.Config) queuejobresources.Interface {
		return NewQueueJobResPod(config)
	})
}

func NewQueueJobResPod(config *rest.Config) queuejobresources.Interface {
	// create k8s clientset

	qjrPod := &QueueJobResPod {
		clients:    kubernetes.NewForConfigOrDie(config),
		arbclients: clientset.NewForConfigOrDie(config),
		deletedPodsCounter: maputils.NewSyncCounterMap(),
		
	}

	// create informer for pod information
	qjrPod.podInformer = informers.NewSharedInformerFactory(qjrPod.clients, 0).Core().V1().Pods()
	qjrPod.podInformer.Informer().AddEventHandler(
		cache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool {
				switch t := obj.(type) {
				case *v1.Pod:
					glog.V(4).Infof("filter pod name(%s) namespace(%s) status(%s)\n", t.Name, t.Namespace, t.Status.Phase)
					return true
				default:
					return false
				}
			},
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc:    qjrPod.addPod,
				UpdateFunc: qjrPod.updatePod,
				DeleteFunc: qjrPod.deletePod,
			},
		})

	qjrPod.rtScheme = runtime.NewScheme()
	v1.AddToScheme(qjrPod.rtScheme)

	qjrPod.jsonSerializer = json.NewYAMLSerializer(json.DefaultMetaFactory, qjrPod.rtScheme, qjrPod.rtScheme)

	qjrPod.podStore = qjrPod.podInformer.Lister()
	qjrPod.podSynced = qjrPod.podInformer.Informer().HasSynced

	qjrPod.refManager = queuejobresources.NewLabelRefManager()

	return qjrPod
}

// Run the main goroutine responsible for watching and pods.
func (qjrPod *QueueJobResPod) Run(stopCh <-chan struct{}) {

	qjrPod.podInformer.Informer().Run(stopCh)
}

func (qjrPod *QueueJobResPod) addPod(obj interface{}) {

	return
}

func (qjrPod *QueueJobResPod) updatePod(old, cur interface{}) {

	return
}


func (cc *QueueJobResPod) deletePod(obj interface{}) {
	var pod *v1.Pod
	switch t := obj.(type) {
	case *v1.Pod:
		pod = t
	case cache.DeletedFinalStateUnknown:
		var ok bool
		pod, ok = t.Obj.(*v1.Pod)
		if !ok {
			glog.Errorf("Cannot convert to *v1.Pod: %v", t.Obj)
			return
		}
	default:
		glog.Errorf("Cannot convert to *v1.Pod: %v", t)
		return
	}

	// update delete pod counter for a QueueJob
	if len(pod.Labels) != 0 && len(pod.Labels[QueueJobNameLabel]) > 0 {
		cc.deletedPodsCounter.DecreaseCounter(fmt.Sprintf("%s/%s", pod.Namespace, pod.Labels[QueueJobNameLabel]))
	}
}


// Parse queue job api object to get Pod template
func (cc *QueueJobResPod) getPodTemplate(qjobRes *arbv1.XQueueJobResource) (*v1.PodTemplateSpec, error) {

	podGVK := schema.GroupVersion{Group: v1.GroupName, Version: "v1"}.WithKind("PodTemplate")

	obj, _, err := cc.jsonSerializer.Decode(qjobRes.Template.Raw, &podGVK, nil)
	if err != nil {
		return nil, err
	}

	template, ok := obj.(*v1.PodTemplate)
	if !ok {
		return nil, fmt.Errorf("Queuejob resource template not define a Pod")
	}

	return &template.Template, nil

}

// filterActivePods returns pods that have not terminated.
func filterActivePods(pods []*v1.Pod) []*v1.Pod {
	var result []*v1.Pod
	for _, p := range pods {
		if isPodActive(p) {
			result = append(result, p)
		} else {
			glog.V(4).Infof("Ignoring inactive pod %v/%v in state %v, deletion time %v",
				p.Namespace, p.Name, p.Status.Phase, p.DeletionTimestamp)
		}
	}
	return result
}

func isPodActive(p *v1.Pod) bool {
	return v1.PodSucceeded != p.Status.Phase &&
		v1.PodFailed != p.Status.Phase &&
		p.DeletionTimestamp == nil
}

func (cc *QueueJobResPod) SyncQueueJob(queuejob *arbv1.XQueueJob, qjobRes *arbv1.XQueueJobResource) error {
	// check if there are still terminating pods for this QueueJob
	counter, ok := cc.deletedPodsCounter.Get(fmt.Sprintf("%s/%s", queuejob.Namespace, queuejob.Name))
	if ok && counter >= 0 {
		return fmt.Errorf("There are still teminating pods for QueueJob %s/%s, can not sync it now", queuejob.Namespace, queuejob.Name)
	}

	pods, err := cc.getPodsForQueueJob(queuejob)
	if err != nil {
		return err
	}
	fmt.Printf("There are %d pods of QueueJob %s\n", len(pods), queuejob.Name)

	err = cc.manageQueueJob(queuejob, pods, qjobRes)
	
	return err
}


// filterPods returns pods based on their phase.
func filterPods(pods []*corev1.Pod, phase corev1.PodPhase) int {
	result := 0
	for i := range pods {
		if phase == pods[i].Status.Phase {
			result++
		}
	}
	return result
}

// manageQueueJob is the core method responsible for managing the number of running
// pods according to what is specified in the job.Spec.
// Does NOT modify <activePods>.
func (cc *QueueJobResPod) manageQueueJob(qj *arbv1.XQueueJob, pods []*v1.Pod, ar *arbv1.XQueueJobResource) error {
	var err error
	replicas := 0
	if qj.Spec.AggrResources.Items != nil {
		// we call clean-up for each controller
		for _, ar := range qj.Spec.AggrResources.Items {
			if ar.Type == arbv1.ResourceTypePod {
				replicas = int(ar.Replicas)
		}
	 }}
	running := int32(filterPods(pods, v1.PodRunning))
	pending := int32(filterPods(pods, v1.PodPending))
	succeeded := int32(filterPods(pods, v1.PodSucceeded))
	failed := int32(filterPods(pods, v1.PodFailed))

	fmt.Printf("There are %d pods of QueueJob %s:  replicas: %d pending %d, running %d, succeeded %d, failed %d",
		len(pods), qj.Name, replicas, pending, running, succeeded, failed)

	ss, err := cc.arbclients.ArbV1().SchedulingSpecs(qj.Namespace).List(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", qj.Name),
	})

	if len(ss.Items) == 0 {
		schedSpc := createQueueJobSchedulingSpec(qj)
		_, err := cc.arbclients.ArbV1().SchedulingSpecs(qj.Namespace).Create(schedSpc)
		if err != nil {
			glog.Errorf("Failed to create SchedulingSpec for QueueJob %v/%v: %v",
				qj.Namespace, qj.Name, err)
		}
	} else {
		glog.V(3).Infof("There's %v SchedulingSpec for QueueJob %v/%v",
			len(ss.Items), qj.Namespace, qj.Name)
	}

	// Create pod if necessary
	if diff := int32(replicas) - pending - running - succeeded; diff > 0 {
		glog.V(3).Infof("Try to create %v Pods for QueueJob %v/%v", diff, qj.Namespace, qj.Name)
		var errs []error
		wait := sync.WaitGroup{}
		wait.Add(int(diff))
		for i := int32(0); i < diff; i++ {
			go func(ix int32) {
				defer wait.Done()
				newPod := cc.createQueueJobPod(qj, ix, ar)
				_, err := cc.clients.Core().Pods(newPod.Namespace).Create(newPod)
				if err != nil {
					// Failed to create Pod, wait a moment and then create it again
					// This is to ensure all pods under the same QueueJob created
					// So gang-scheduling could schedule the QueueJob successfully
					glog.Errorf("Failed to create pod %s for QueueJob %s, err %#v",
						newPod.Name, qj.Name, err)
					errs = append(errs, err)
				}
			}(i)
		}
		wait.Wait()

		if len(errs) != 0 {
			return fmt.Errorf("failed to create %d pods of %d", len(errs), diff)
		}
	}
	qj.Status = arbv1.XQueueJobStatus {
		Pending:      pending,
		Running:      running,
		Succeeded:    succeeded,
		Failed:       failed,
		MinAvailable: int32(qj.Spec.SchedSpec.MinAvailable),
	}

	return err
}

func (cc *QueueJobResPod) getPodsForQueueJob(qj *arbv1.XQueueJob) ([]*v1.Pod, error) {
	sel := &metav1.LabelSelector{
				MatchLabels: map[string]string{
					queueJobName: qj.Name,
				},
			}
	selector, err := metav1.LabelSelectorAsSelector(sel)
	if err != nil {
		return nil, fmt.Errorf("couldn't convert QueueJob selector: %v", err)
	}

	// List all pods under QueueJob
	pods, errt := cc.podStore.Pods(qj.Namespace).List(selector)
	if errt != nil {
		return nil, errt
	}

	return pods, nil
}

// manageQueueJobPods is the core method responsible for managing the number of running
// pods according to what is specified in the job.Spec. This is a controller for all pods specified in the QJ template
// Does NOT modify <activePods>.
func (cc *QueueJobResPod) manageQueueJobPods(activePods []*v1.Pod, succeeded int32, qj *arbv1.XQueueJob, ar *arbv1.XQueueJobResource) (bool, error) {
	jobDone := false
	var err error
	active := int32(len(activePods))

	replicas := 0
        if qj.Spec.AggrResources.Items != nil {
                // we call clean-up for each controller
                for _, ar := range qj.Spec.AggrResources.Items {
                        if ar.Type == arbv1.ResourceTypePod {
                                replicas = replicas + 1
                }
         }}

	if active+succeeded > int32(replicas) {
		// the QueueJob replicas is reduce by user, terminated all pods for gang scheduling
		// and re-create pods for the queuejob in next loop
		jobDone = false
		// TODO(jinzhejz): need make sure manage this QueueJob after all old pods are terminated
		err = cc.terminatePodsForQueueJob(qj)
	} else if active+succeeded == int32(replicas) {
		// pod number match QueueJob replicas perfectly
		if succeeded == int32(replicas) {
			// all pods exit successfully
			jobDone = true
		} else {
			// some pods are still running
			jobDone = false
		}
	} else if active+succeeded < int32(replicas) {
		if active+succeeded == 0 {
			// it is a new QueueJob, create pods for it
			diff := int32(replicas) - active - succeeded

			wait := sync.WaitGroup{}
			wait.Add(int(diff))
			for i := int32(0); i < diff; i++ {
				go func(ix int32) {
					defer wait.Done()
					newPod := cc.createQueueJobPod(qj, ix, ar)
					//newPod := buildPod(fmt.Sprintf("%s-%d-%s", qj.Name, ix, generateUUID()), qj.Namespace, qj.Spec.Template, []metav1.OwnerReference{*metav1.NewControllerRef(qj, controllerKind)}, ix)
					for {
						_, err := cc.clients.Core().Pods(newPod.Namespace).Create(newPod)
						if err == nil {
							// Create Pod successfully
							break
						} else {
							// Failed to create Pod, wait a moment and then create it again
							// This is to ensure all pods under the same QueueJob created
							// So gang-scheduling could schedule the QueueJob successfully
							glog.Warningf("Failed to create pod %s for QueueJob %s, err %#v, wait 2 seconds and re-create it", newPod.Name, qj.Name, err)
							time.Sleep(2 * time.Second)
						}
					}
				}(i)
			}
			wait.Wait()
			jobDone = false
		} else if active+succeeded > 0 {
			// the QueueJob replicas is reduce by user, terminated all pods for gang scheduling
			// and re-create pods for the queuejob in next loop
			jobDone = false
			// TODO(jinzhejz): need make sure manage this QueueJob after all old pods are terminated
			err = cc.terminatePodsForQueueJob(qj)
		}
	}

	return jobDone, err
}

func (cc *QueueJobResPod) terminatePodsForQueueJob(qj *arbv1.XQueueJob) error {
	pods, err := cc.getPodsForQueueJob(qj)
	if len(pods) == 0 || err != nil {
		return err
	}

	cc.deletedPodsCounter.Set(fmt.Sprintf("%s/%s", qj.Namespace, qj.Name), len(pods))

	wait := sync.WaitGroup{}
	wait.Add(len(pods))
	for _, pod := range pods {
		go func(p *v1.Pod) {
			defer wait.Done()
			err := cc.clients.Core().Pods(p.Namespace).Delete(p.Name, &metav1.DeleteOptions{})
			if err != nil {
				glog.Warning("Fail to delete pod %s for QueueJob %s/%s", p.Name, qj.Namespace, qj.Name)
				cc.deletedPodsCounter.DecreaseCounter(fmt.Sprintf("%s/%s", qj.Namespace, qj.Name))
			}
		}(pod)
	}
	wait.Wait()

	return nil
}


func (qjrPod *QueueJobResPod) getPodsForQueueJobRes(qjobRes *arbv1.XQueueJobResource, j *arbv1.XQueueJob) ([]*v1.Pod, error) {

	pods, err := qjrPod.getPodsForQueueJob(j)
	if err != nil {
		return nil, err
	}

	myPods := []*v1.Pod{}
	for i, pod := range pods {
		if qjrPod.refManager.BelongTo(qjobRes, pod) {
			myPods = append(myPods, pods[i])
		}
	}

	return myPods, nil

}

func generateUUID() string {
	id := uuid.NewV1()

	return fmt.Sprintf("%s", id)
}

func (qjrPod *QueueJobResPod) deleteQueueJobResPods(qjobRes *arbv1.XQueueJobResource, queuejob *arbv1.XQueueJob) error {

	job := *queuejob

	pods, err := qjrPod.getPodsForQueueJobRes(qjobRes, queuejob)
	if err != nil {
		return err
	}

	activePods := filterActivePods(pods)
	active := int32(len(activePods))

	wait := sync.WaitGroup{}
	wait.Add(int(active))
	for i := int32(0); i < active; i++ {
		go func(ix int32) {
			defer wait.Done()
			if err := qjrPod.clients.Core().Pods(queuejob.Namespace).Delete(activePods[ix].Name, &metav1.DeleteOptions{}); err != nil {
				defer utilruntime.HandleError(err)
				glog.V(2).Infof("Failed to delete %v, queue job %q/%q deadline exceeded", activePods[ix].Name, job.Namespace, job.Name)
			}
		}(i)
	}
	wait.Wait()

	return nil
}


func createQueueJobSchedulingSpec(qj *arbv1.XQueueJob) *arbv1.SchedulingSpec {
        return &arbv1.SchedulingSpec{
                ObjectMeta: metav1.ObjectMeta{
                        Name:      qj.Name,
                        Namespace: qj.Namespace,
                        OwnerReferences: []metav1.OwnerReference{
                                *metav1.NewControllerRef(qj, queueJobKind),
                        },
                },
                Spec: qj.Spec.SchedSpec,
        }
}

func (qjrPod *QueueJobResPod) createQueueJobPod(qj *arbv1.XQueueJob, ix int32, qjobRes *arbv1.XQueueJobResource) *corev1.Pod {
        templateCopy, err := qjrPod.getPodTemplate(qjobRes)

	if err != nil {
		glog.Errorf("Cannot parse pod template for QJ")
		return nil
	}
        podName := fmt.Sprintf("%s-%d-%s", qj.Name, ix, generateUUID())

	tmpl := templateCopy.Labels
	tmpl[queueJobName] = qj.Name
        return &corev1.Pod{
                ObjectMeta: metav1.ObjectMeta{
                        Name:      podName,
                        Namespace: qj.Namespace,
                        OwnerReferences: []metav1.OwnerReference{
                                *metav1.NewControllerRef(qj, queueJobKind),
                        },
                        Labels: tmpl,
                },
                Spec: templateCopy.Spec,
        }
}
   

func (qjrPod *QueueJobResPod) Cleanup(queuejob *arbv1.XQueueJob, qjobRes *arbv1.XQueueJobResource) error {

	return qjrPod.deleteQueueJobResPods(qjobRes, queuejob)
}

