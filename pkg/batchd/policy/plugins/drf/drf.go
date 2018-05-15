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

package drf

import (
	"github.com/kubernetes-incubator/kube-arbitrator/pkg/batchd/policy/api"
	"github.com/kubernetes-incubator/kube-arbitrator/pkg/batchd/policy/framework"
)

type drfAttr struct {
	share            float64
	dominantResource string
	allocated        *api.Resource
}

type drfPlugin struct {
	totalResource *api.Resource

	// Key is Request ID
	requestOpts map[string]*drfAttr
}

func New() *drfPlugin {
	return &drfPlugin{
		totalResource: api.EmptyResource(),
		requestOpts:   map[string]*drfAttr{},
	}
}

func (drf *drfPlugin) Name() string {
	return "drf"
}

func (drf *drfPlugin) OnSessionEnter(ssn *framework.Session) {
	// Prepare scheduling data for this session.
	for _, n := range ssn.Nodes {
		drf.totalResource.Add(n.Allocatable)
	}

	for _, req := range ssn.Requests {
		attr := &drfAttr{
			allocated: api.EmptyResource(),
		}

		for _, ru := range req.Units {
			if api.OccupiedResource(ru.Status) {
				attr.allocated.Add(ru.Resreq)
			}
		}

		drf.requestOpts[req.ID] = attr
	}

	// Register Request Order function.
	ssn.RegisterRequestOrderFn(func(l interface{}, r interface{}) bool {
		lv := l.(*api.Request)
		rv := r.(*api.Request)

		return drf.requestOpts[lv.ID].share < drf.requestOpts[rv.ID].share
	})

	// Register event handlers.
	ssn.AddEventHandler(&framework.EventHandlers{
		OnAllocated: func(event *framework.Event) error {
			attr := drf.requestOpts[event.RequestID]
			attr.allocated.Add(event.RequestUnit.Resreq)

			// Refresh share of Request
			attr.share = 0
			for _, rn := range api.ResourceNames {
				share := attr.allocated.Get(rn) / drf.totalResource.Get(rn)
				if share > attr.share {
					attr.share = share
				}
			}

			return nil
		},
	})
}

func (drf *drfPlugin) OnSessionLeave(session *framework.Session) {
	// Clean schedule data.
	drf.totalResource = api.EmptyResource()
	drf.requestOpts = map[string]*drfAttr{}
}
