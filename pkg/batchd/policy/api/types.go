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

import "time"

type RequestStatus int

const (
	// The request unit is pending for scheduling
	Pending RequestStatus = 1 << iota
	// The request unit is allocated, but scheduling did not submit
	// Bind request to meta-database.
	Allocated
	// The request unit is preempted by the scheduler, but the preempted request
	// is not submitted.
	Preempted
	// The Bind request is sent to meta-database, waiting for the
	// feedback.
	Binding
	// The scheduler/allocator get the Bind feedback.
	Bound
	// There's some 'task' are running on this allocation.
	Running
	Releasing
	Failed
	Succeeded
	Unknown
)

func OccupiedResource(rs RequestStatus) bool {
	switch rs {
	case Allocated, Binding, Bound, Running:
		return true
	}
	return false
}

type RequestUnit struct {
	ID string
	// The id of request that belong to
	RequestID string
	// The status of task on this unit
	Status RequestStatus
	// Resource request of this unit
	Resreq *Resource
	// The priority of request unit
	Priority int32
	// The host that allocated to this unit
	Host string
	// Candidate host list for this unit
	Candidates []string
}

type Request struct {
	ID                string
	Consumer          string
	User              string
	Units             map[string]*RequestUnit
	CreationTimestamp time.Time
}

type Node struct {
	ID          string
	Name        string
	Allocatable *Resource
	Idle        *Resource
	Units       map[string]*RequestUnit
}

type DecisionType int

const (
	Allocate DecisionType = 1 << iota
	Terminate
)

type Decision struct {
	RequestID     string
	RequestUnitID string
	Host          string
	Type          DecisionType
}

//
//type Decisions struct {
//	ID           string
//	Host         string
//	Allocations  []*RequestUnit
//	Terminations []*RequestUnit
//}
