package gang

import (
	"fmt"
	"github.com/kubernetes-sigs/kube-batch/pkg/scheduler/api"
	"k8s.io/apimachinery/pkg/types"
	"testing"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestJobReady(t *testing.T) {
	tests := []struct {
		name     string
		job      *api.JobInfo
		expected bool
	}{
		{
			name:     "ready job",
			job:      buildJob("ready-job", api.Ready),
			expected: true,
		},
		{
			name:     "almost ready job",
			job:      buildJob("almost-ready-job", api.OverResourceReady),
			expected: false,
		},
		{
			name:     "not job",
			job:      buildJob("partially-allocated-job", api.NotReady),
			expected: false,
		},
	}

	for i, test := range tests {
		actual := jobReady(test.job)
		if test.expected != actual {
			t.Errorf("case %d (%s): expected: %v, got %v ", i, test.name, test.expected, actual)
		}
	}
}

func buildPod(ns, n, nn string, p v1.PodPhase, req v1.ResourceList, owner []metav1.OwnerReference, labels map[string]string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:             types.UID(fmt.Sprintf("%v-%v", ns, n)),
			Name:            n,
			Namespace:       ns,
			OwnerReferences: owner,
			Labels:          labels,
		},
		Status: v1.PodStatus{
			Phase: p,
		},
		Spec: v1.PodSpec{
			NodeName: nn,
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Requests: req,
					},
				},
			},
			Priority: new(int32),
		},
	}
}

func buildResourceList(cpu string, memory string) v1.ResourceList {
	return v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse(cpu),
		v1.ResourceMemory: resource.MustParse(memory),
	}
}

func buildTask(name string, status api.TaskStatus) *api.TaskInfo {
	pod := buildPod("c1", name, "n1", v1.PodPending, buildResourceList("1000m", "1G"), []metav1.OwnerReference{}, make(map[string]string))
	task := api.NewTaskInfo(pod)
	task.Status = status
	return task
}

func buildJob(name string, status api.JobReadiness) *api.JobInfo {
	var t1, t2 *api.TaskInfo

	job := api.NewJobInfo(api.JobID(name))
	job.MinAvailable = 2

	switch status {
	case api.Ready:
		t1 = buildTask("t1", api.Allocated)
		t2 = buildTask("t2", api.Allocated)
	case api.OverResourceReady:
		t1 = buildTask("t1", api.Allocated)
		t2 = buildTask("t2", api.AllocatedOverBackfill)
	case api.NotReady:
	}

	if t1 != nil {
		job.AddTaskInfo(t1)
	}

	if t2 != nil {
		job.AddTaskInfo(t2)
	}

	return job
}
