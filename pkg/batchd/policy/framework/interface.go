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

// Action is the interface of actions.
type Action interface {
	// The unique name of action.
	Name() string

	// Initialize initializes the allocator action.
	Initialize()

	// Execute executes the action for resource allocation of cluster.
	Execute(ssn *Session)

	// UnIntialize un-initializes the allocator action.
	UnInitialize()
}

type Plugin interface {
	// The name of plugin
	Name() string

	// OnSessionEnter is the callback func when framework opens a session
	OnSessionEnter(session *Session)

	// OnSessionLeave is the callback func when framework closes a session
	OnSessionLeave(session *Session)
}
