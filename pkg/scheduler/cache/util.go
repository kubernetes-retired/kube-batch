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
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kubernetes-sigs/kube-batch/pkg/apis/scheduling/v1alpha1"
	kbv1 "github.com/kubernetes-sigs/kube-batch/pkg/apis/scheduling/v1alpha1"
	"github.com/kubernetes-sigs/kube-batch/pkg/apis/utils"
	"github.com/kubernetes-sigs/kube-batch/pkg/scheduler/api"
)

const (
	shadowPodGroupKey = "kube-batch/shadow-pod-group"
)

func shadowPodGroup(pg *v1alpha1.PodGroup) bool {
	if pg == nil {
		return true
	}

	_, found := pg.Annotations[shadowPodGroupKey]

	return found
}

func createShadowPodGroup(pod *v1.Pod) *v1alpha1.PodGroup {
	jobID := api.JobID(utils.GetController(pod))
	if len(jobID) == 0 {
		jobID = api.JobID(pod.UID)
	}

	return &v1alpha1.PodGroup{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: pod.Namespace,
			Name:      string(jobID),
			Annotations: map[string]string{
				shadowPodGroupKey: string(jobID),
			},
		},
		Spec: v1alpha1.PodGroupSpec{
			MinMember: 1,
		},
	}
}

func GenerateJob(mpi metav1.Object, job batchv1.Job) *batchv1.Job {
	// TODO: Default and Validate these with Webhooks
	copyLabels := job.GetLabels()
	if copyLabels == nil {
		copyLabels = map[string]string{}
	}
	labels := map[string]string{}
	for k, v := range copyLabels {
		labels[k] = v
	}
	labels["job-name"] = mpi.GetName()
	labels["job-type"] = "mpi-job"

	job.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: labels,
	}

	copyPodAnnotations := job.Spec.Template.GetAnnotations()
	if copyPodAnnotations == nil {
		copyPodAnnotations = map[string]string{}
	}
	podAnnotations := map[string]string{}
	for k, v := range copyPodAnnotations {
		podAnnotations[k] = v
	}
	podAnnotations["scheduling.k8s.io/group-name"] = mpi.GetName() + "-podgroup"

	hostfileMount := corev1.VolumeMount{
		Name:      "mpi-hostfile",
		ReadOnly:  true,
		MountPath: "/mpi/",
	}
	mpiMount := corev1.VolumeMount{
		Name:      "mpi-data",
		ReadOnly:  true,
		MountPath: "/entry/",
	}

	copyPodContainers := job.Spec.Template.Spec.Containers
	containers := []corev1.Container{}
	for _, v := range copyPodContainers {
		newArgs := []string{}
		v.VolumeMounts = append(v.VolumeMounts, hostfileMount)
		v.VolumeMounts = append(v.VolumeMounts, mpiMount)
		for _, cmd := range v.Command {
			newArgs = append(newArgs, cmd)
		}
		for _, arg := range v.Args {
			newArgs = append(newArgs, arg)
		}

		v.Command = []string{"/entry/startup.sh"}
		v.Args = newArgs

		v.Env = append(v.Env,
			corev1.EnvVar{
				Name:  "HYDRA_BOOTSTRAP",
				Value: "rsh",
			},
			corev1.EnvVar{
				Name:  "HYDRA_BOOTSTRAP_EXEC",
				Value: "/entry/kubeexec.sh",
			},
			corev1.EnvVar{
				Name:  "HYDRA_HOST_FILE",
				Value: "/mpi/hostfile",
			})

		containers = append(containers, v)
	}

	defaultMode := int32(0444)
	scriptMode := int32(0777)
	job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, corev1.Volume{
		Name: "mpi-hostfile",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: mpi.GetName() + "-hostfile",
				},
				Items: []corev1.KeyToPath{
					{
						Key:  "hostfile",
						Path: "hostfile",
						Mode: &defaultMode,
					},
				},
			},
		},
	})
	job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, corev1.Volume{
		Name: "mpi-data",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: mpi.GetName() + "-mpi-data",
				},
				Items: []corev1.KeyToPath{
					{
						Key:  "kubeexec.sh",
						Path: "kubeexec.sh",
						Mode: &scriptMode,
					},
					{
						Key:  "startup.sh",
						Path: "startup.sh",
						Mode: &scriptMode,
					},
					{
						Key:  "executor",
						Path: "executor",
						Mode: &defaultMode,
					},
					{
						Key:  "hosts",
						Path: "hosts",
						Mode: &defaultMode,
					},
				},
			},
		},
	})

	theTrueTrue := true

	job.Spec.ManualSelector = &theTrueTrue
	job.Spec.Template.Labels = labels
	job.Spec.Template.Annotations = podAnnotations
	job.Spec.Template.Spec.ServiceAccountName = mpi.GetName() + "-sa"
	job.Spec.Template.Spec.RestartPolicy = "Never"
	job.Spec.Template.Spec.SchedulerName = "kube-batch"
	job.Spec.Template.Spec.Containers = containers

	gvk := schema.GroupVersionKind{
		Group:   kbv1.GroupName,
		Version: kbv1.GroupVersion,
		Kind:    "MPI",
	}

	newjob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:         mpi.GetName(),
			GenerateName: mpi.GetName() + "-",
			Namespace:    mpi.GetNamespace(),
			Labels:       labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mpi, gvk),
			},
		},
		Spec: job.Spec,
	}
	return newjob
}
