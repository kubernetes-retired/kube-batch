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

package api

// TaskStatus defines the status of a task/pod.
type TaskStatus int

const (
	// Pending means the task is pending in the apiserver.
	Pending TaskStatus = 1 << iota

	// Borrowing means that the task is allocated borrowing resources
	// that are currently occupied by tasks.
	// Task T is allocated to Node N as an Borrowing task if,
	// and only if, N.IdleResource < T.RequestResource <= N.AllocatableResource.
	Borrowing

	// Allocated means the scheduler assigns a host to it.
	Allocated

	// Pipelined means the scheduler assigns a host to wait for releasing resource.
	Pipelined

	// Binding means the scheduler send Bind request to apiserver.
	Binding

	// Bound means the task/Pod bounds to a host.
	Bound

	// Running means a task is running on the host.
	Running

	// Releasing means a task/pod is deleted.
	Releasing

	// Succeeded means that all containers in the pod have voluntarily terminated
	// with a container exit code of 0, and the system is not going to restart any of these containers.
	Succeeded

	// Failed means that all containers in the pod have terminated, and at least one container has
	// terminated in a failure (exited with a non-zero exit code or was stopped by the system).
	Failed

	// Unknown means the status of task/pod is unknown to the scheduler.
	Unknown
)

type TaskCondition struct {
	IsBackfill bool
}

// JobReadiness type of job readiness
type JobReadiness int

const (
	// Ready : a job is Ready if the number of tasks in Allocated state
	// exceeds the job's minimum task number requirement. In other words,
	// #(Allocated Tasks) >= Job.MinAvailable
	// A Ready job can be dispatch to a node right away.
	Ready JobReadiness = 1 << iota

	// ConditionallyReady : a job is ConditionallyReady if the job is not Ready for dispatch, but
	// the number of tasks in Allocated state exceeds the job's minim task
	// number requirement. In other words,
	// #(Allocated Tasks) < Job.MinAvailable &&
	// #(Allocated Tasks) + #(AllocatedOverBackFill Tasks) >= Job.MinAvailable
	ConditionallyReady

	// NotReady : #(Allocated Tasks) + #(AllocatedOverBackFill Tasks) < Job.MinAvailable
	NotReady
)

// AllocatedStatuses all status of allocated
func AllocatedStatuses() []TaskStatus {
	return []TaskStatus{Bound, Binding, Running, Allocated}
}

func (ts TaskStatus) String() string {
	switch ts {
	case Pending:
		return "Pending"
	case Allocated:
		return "Allocated"
	case Borrowing:
		return "Borrowing"
	case Binding:
		return "Binding"
	case Bound:
		return "Bound"
	case Running:
		return "Running"
	case Releasing:
		return "Releasing"
	case Succeeded:
		return "Succeeded"
	case Failed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// validateStatusUpdate validates whether the status transfer is valid.
func validateStatusUpdate(oldStatus, newStatus TaskStatus) error {
	return nil
}

// LessFn is the func declaration used by sort or priority queue.
type LessFn func(interface{}, interface{}) bool

// CompareFn is the func declaration used by sort or priority queue.
type CompareFn func(interface{}, interface{}) int

// ValidateFn is the func declaration used to check object's status.
type ValidateFn func(interface{}) bool

// ValidateResult is struct to which can used to determine the result
type ValidateResult struct {
	Pass    bool
	Reason  string
	Message string
}

// ValidateExFn is the func declaration used to validate the result
type ValidateExFn func(interface{}) *ValidateResult

// PredicateFn is the func declaration used to predicate node for task.
type PredicateFn func(*TaskInfo, *NodeInfo) error

// EvictableFn is the func declaration used to evict tasks.
type EvictableFn func(*TaskInfo, []*TaskInfo) []*TaskInfo

// NodeOrderFn is the func declaration used to get priority score for a node for a particular task.
type NodeOrderFn func(*TaskInfo, *NodeInfo) (float64, error)

// BackFillEligibleFn is the func declaration used to get backfill eligiblity
type BackFillEligibleFn func(interface{}) bool
