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
	"sync"

	"k8s.io/client-go/rest"

	schedcache "github.com/kubernetes-incubator/kube-arbitrator/pkg/batchd/cache"
	policyapi "github.com/kubernetes-incubator/kube-arbitrator/pkg/batchd/policy/api"
)

var mutex sync.Mutex
var clusterCache schedcache.Cache
var plugins []Plugin

func Run(config *rest.Config, schedulerName string, stopCh <-chan struct{}) {
	clusterCache = schedcache.New(config, schedulerName)

	// Start cache for policy.
	go clusterCache.Run(stopCh)
	clusterCache.WaitForCacheSync(stopCh)
}

func OpenSession() *Session {
	ssn := &Session{
		Requests: map[string]*policyapi.Request{},
		Nodes:    map[string]*policyapi.Node{},
	}

	reqs, nodes := clusterCache.Snapshot()

	for _, req := range reqs {
		ssn.Requests[req.ID] = req
	}

	for _, n := range nodes {
		ssn.Nodes[n.ID] = n
	}

	for _, p := range plugins {
		p.OnSessionEnter(ssn)
	}

	return ssn
}

func CloseSession(ssn *Session) {
	for _, p := range plugins {
		p.OnSessionLeave(ssn)
	}

	// Cleanup scheduling data in session.
	ssn.Requests = nil
	ssn.Backoff = nil
	ssn.Nodes = nil
}

func RegisterPlugin(p Plugin) error {
	mutex.Lock()
	defer mutex.Unlock()

	plugins = append(plugins, p)
	return nil
}
