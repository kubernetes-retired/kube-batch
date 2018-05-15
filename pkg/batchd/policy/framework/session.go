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

package framework

import (
	policyapi "github.com/kubernetes-incubator/kube-arbitrator/pkg/batchd/policy/api"
)

type Event struct {
	RequestID     string
	RequestUnitID string
	Host          string
	RequestUnit   *policyapi.RequestUnit
}

type EventHandler func(event *Event) error

type EventHandlers struct {
	OnAllocated EventHandler
}

type Session struct {
	ID       string
	Requests map[string]*policyapi.Request
	Nodes    map[string]*policyapi.Node

	Backoff map[string]*policyapi.Request

	RequestOrderFn     policyapi.LessFn
	RequestUnitOrderFn policyapi.LessFn

	EventHandlers *EventHandlers
}

type decisionHandler func(ssn *Session, ds *policyapi.Decision) error

var decisionHandlers = map[policyapi.DecisionType]decisionHandler{
	policyapi.Allocate: allocateDecisionHandler,
}

func (ssn *Session) PublishDecision(ds *policyapi.Decision) error {
	return (decisionHandlers[ds.Type])(ssn, ds)
}

func (ssn *Session) AddEventHandler(eh *EventHandlers) {
	ssn.EventHandlers = eh
}

func (ssn *Session) RegisterRequestOrderFn(fn policyapi.LessFn) {
	ssn.RequestOrderFn = fn
}

func (ssn *Session) BackoffRequest(reqID string) {
	req := ssn.Requests[reqID]
	ssn.Backoff[reqID] = req
	delete(ssn.Requests, reqID)
}

func allocateDecisionHandler(ssn *Session, ds *policyapi.Decision) error {
	ru := ssn.Requests[ds.RequestID].Units[ds.RequestUnitID]
	node := ssn.Nodes[ds.Host]

	ru.Status = policyapi.Allocated
	ru.Host = ds.Host
	node.Idle.Sub(ru.Resreq)

	node.Units[ru.ID] = ru

	if ssn.EventHandlers != nil && ssn.EventHandlers.OnAllocated != nil {
		ssn.EventHandlers.OnAllocated(&Event{
			RequestUnitID: ru.ID,
			RequestID:     ds.RequestID,
			Host:          ds.Host,
			RequestUnit:   ru,
		})
	}

	return nil
}
