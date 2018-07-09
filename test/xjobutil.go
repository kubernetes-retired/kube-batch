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

package test

import (
	json2 "encoding/json"
	"fmt"
	. "github.com/onsi/gomega"
	"time"

	"k8s.io/apimachinery/pkg/runtime"

	arbv1 "github.com/kubernetes-incubator/kube-arbitrator/pkg/apis/v1alpha1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
)


func createXQueueJob(context *context, name string, min, rep int32, priority string, img string, req v1.ResourceList) *arbv1.XQueueJob {
	queueJobName := "xqueuejob.arbitrator.k8s.io"

	podTemplate := v1.PodTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{queueJobName: name},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "PodTemplate"},
		Template: v1.PodTemplateSpec{
			Spec: v1.PodSpec{

				PriorityClassName: priority,
				RestartPolicy:     v1.RestartPolicyNever,

				Containers: []v1.Container{
					{
						Image:           img,
						Name:            name,
						ImagePullPolicy: v1.PullIfNotPresent,
						Resources: v1.ResourceRequirements{
							Requests: req,
						},
					},
				},
			}},
	}

	pods := make([]arbv1.XQueueJobResource, 0)

	data, err := json2.Marshal(podTemplate)
	if err != nil {
		fmt.Errorf("I encode podTemplate %+v %+v", podTemplate, err)
	}

	rawExtension := runtime.RawExtension{Raw: json2.RawMessage(data)}

	podResource := arbv1.XQueueJobResource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: context.namespace,
			Labels:    map[string]string{queueJobName: name},
		},
		Replicas:          rep,
		MinAvailable:      &min,
		AllocatedReplicas: 0,
		Priority:          0.0,
		Type:              arbv1.ResourceTypePod,
		Template:          rawExtension,
	}

	pods = append(pods, podResource)

	queueJob := &arbv1.XQueueJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: context.namespace,
			Labels:    map[string]string{queueJobName: name},
		},
		Spec: arbv1.XQueueJobSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					queueJobName: name,
				},
			},
			SchedSpec: arbv1.SchedulingSpecTemplate{
				MinAvailable: int(min),
			},
			AggrResources: arbv1.XQueueJobResourceList{
				Items: pods,
			},
		},
	}

	queueJob, err2 := context.karclient.ArbV1().XQueueJobs(context.namespace).Create(queueJob)
	Expect(err2).NotTo(HaveOccurred())

	return queueJob
}

func xtaskReady(ctx *context, jobName string, taskNum int) wait.ConditionFunc {
	return func() (bool, error) {
		queueJob, err := ctx.karclient.ArbV1().XQueueJobs(ctx.namespace).Get(jobName, metav1.GetOptions{})
		pods, err := ctx.kubeclient.CoreV1().Pods(ctx.namespace).List(metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())

		labelSelector := labels.SelectorFromSet(queueJob.Spec.Selector.MatchLabels)

		readyTaskNum := 0
		for _, pod := range pods.Items {
			if !labelSelector.Matches(labels.Set(pod.Labels)) {
				continue
			}
			if pod.Status.Phase == v1.PodRunning || pod.Status.Phase == v1.PodSucceeded {
				readyTaskNum++
			}
		}

		if taskNum < 0 {
			taskNum = queueJob.Spec.SchedSpec.MinAvailable
		}

		return taskNum <= readyTaskNum, nil
	}
}

func listXQueueJob(ctx *context, jobName string) error {
	_, err := ctx.karclient.ArbV1().XQueueJobs(ctx.namespace).Get(jobName, metav1.GetOptions{})
	return err
}

func listXTasks(ctx *context, nJobs int) wait.ConditionFunc {
	return func() (bool, error) {
		jobs, err := ctx.karclient.ArbV1().XQueueJobs(ctx.namespace).List(metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())

		nJobs0 := len(jobs.Items)

		return nJobs0 == nJobs, nil
	}
}

func xtaskCreated(ctx *context, jobName string, taskNum int) wait.ConditionFunc {
	return func() (bool, error) {
		_, err := ctx.karclient.ArbV1().XQueueJobs(ctx.namespace).Get(jobName, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		return true, nil
	}
}

func listXQueueJobs(ctx *context, nJobs int) error {
	return wait.Poll(100*time.Millisecond, oneMinute, listXTasks(ctx, nJobs))
}

func deleteXQueueJob(ctx *context, jobName string) error {
	foreground := metav1.DeletePropagationForeground
	return ctx.karclient.ArbV1().XQueueJobs(ctx.namespace).Delete(jobName, &metav1.DeleteOptions{
		PropagationPolicy: &foreground,
	})
}

func waitXJobReady(ctx *context, name string, taskNum int) error {
	return wait.Poll(100*time.Millisecond, oneMinute, xtaskReady(ctx, name, taskNum))
}

func waitXJobCreated(ctx *context, name string) error {
	return wait.Poll(100*time.Millisecond, oneMinute, xtaskCreated(ctx, name, -1))
}

func xjobNotReady(ctx *context, jobName string) wait.ConditionFunc {
	return func() (bool, error) {
		queueJob, err := ctx.karclient.ArbV1().XQueueJobs(ctx.namespace).Get(jobName, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		pods, err := ctx.kubeclient.CoreV1().Pods(ctx.namespace).List(metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())

		labelSelector := labels.SelectorFromSet(queueJob.Spec.Selector.MatchLabels)

		pendingTaskNum := int32(0)
		for _, pod := range pods.Items {
			if !labelSelector.Matches(labels.Set(pod.Labels)) {
				continue
			}
			if pod.Status.Phase == v1.PodPending && len(pod.Spec.NodeName) == 0 {
				pendingTaskNum++
			}
		}

		return int(pendingTaskNum) >= int(queueJob.Spec.SchedSpec.MinAvailable), nil
	}
}

func waitXJobNotReady(ctx *context, name string) error {
	return wait.Poll(10*time.Second, oneMinute, xjobNotReady(ctx, name))
}
