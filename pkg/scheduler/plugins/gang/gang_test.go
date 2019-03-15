package gang

import (
	"github.com/kubernetes-sigs/kube-batch/pkg/scheduler/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestJobOrderFn(t *testing.T) {
	creationTime := metav1.Now()

	readyJob := buildJob(api.JobID("ready-job"), 1, metav1.Now())
	readyJob.AddTaskInfo(buildTask(api.TaskID("running-task"), readyJob.UID, api.Running))

	nonReadyJob := buildJob(api.JobID("unready-job"), 1, creationTime)
	nonReadyJob.AddTaskInfo(buildTask(api.TaskID("pending-task"), nonReadyJob.UID, api.Pending))

	nonReadyJobSameCreationTime := buildJob(api.JobID("unready-job-same-creation-time"), 1, creationTime)
	nonReadyJobSameCreationTime.AddTaskInfo(buildTask(api.TaskID("pending-task"), nonReadyJobSameCreationTime.UID, api.Pending))

	nonReadyJobLaterCreationTime := buildJob(api.JobID("unready-job-later-creation-time"), 1, metav1.Now())
	nonReadyJobLaterCreationTime.AddTaskInfo(buildTask(api.TaskID("pending-task"), nonReadyJobLaterCreationTime.UID, api.Pending))

	tests := []struct{
		name string
		l *api.JobInfo
		r *api.JobInfo
		expected int
	} {
		{
			name: "Ready job is greater than unready job",
			l: readyJob,
			r: nonReadyJob,
			expected: 1,
		},
		{
			name: "Unready job is less than ready job",
			l: nonReadyJob,
			r: readyJob,
			expected: -1,
		},
		{
			name: "Ready jobs are equal",
			l: readyJob,
			r: readyJob,
			expected: 0,
		},
		{
			name: "An unready job should be equal to itself",
			l: nonReadyJob,
			r: nonReadyJob,
			expected: 0,
		},
		{
			name: "Unready jobs are equal",
			l: nonReadyJob,
			r: nonReadyJobLaterCreationTime,
			expected: 0,
		},
	}

	for _, tc := range tests {
		actual := jobOrderFn(tc.l, tc.r)
		if tc.expected != actual {
			t.Errorf("expected: %v, got: %v. Test case: %s, l-value: '%v', r-value: '%v'", tc.expected, actual, tc.name, tc.l, tc.r)
		}
	}
}

func buildJob(jid api.JobID, minAvailable int32, creationTime metav1.Time) *api.JobInfo {
	job := api.NewJobInfo(jid)
	job.MinAvailable = minAvailable
	job.CreationTimestamp = creationTime
	return job
}

func buildTask(tid api.TaskID, jid api.JobID, status api.TaskStatus) *api.TaskInfo {
	return 	&api.TaskInfo{
		UID:       tid,
		Job:       jid,
		Resreq:    api.EmptyResource(),
		Status:    status,
	}
}