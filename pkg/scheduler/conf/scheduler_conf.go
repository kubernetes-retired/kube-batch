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

package conf

import "time"

// SchedulerConfiguration defines the configuration of scheduler.
type SchedulerConfiguration struct {
	// Actions defines the actions list of scheduler in order
	Actions []SchedulerAction `yaml:"actions"`
	// Tiers defines plugins in different tiers
	Tiers []Tier `yaml:"tiers"`
}

const (
	// DefaultStarvingThreshold sets the default value for Starvation Threshold
	DefaultStarvingThreshold time.Duration = 48 * time.Hour
)

// Tier defines plugin tier
type Tier struct {
	Plugins []PluginOption `yaml:"plugins"`
}

// SchedulerAction defines action with its options
type SchedulerAction struct {
	Name      string            `yaml:"name"`
	Arguments map[string]string `yaml:"arguments"`
}

// PluginOption defines the options of plugin
type PluginOption struct {
	// The name of Plugin
	Name string `yaml:"name"`
	// EnabledJobOrder defines whether jobOrderFn is enabled
	EnabledJobOrder *bool `yaml:"enableJobOrder"`
	// EnabledJobReady defines whether jobReadyFn is enabled
	EnabledJobReady *bool `yaml:"enableJobReady"`
	// EnabledJobPipelined defines whether jobPipelinedFn is enabled
	EnabledJobPipelined *bool `yaml:"enableJobPipelined"`
	// EnabledTaskOrder defines whether taskOrderFn is enabled
	EnabledTaskOrder *bool `yaml:"enableTaskOrder"`
	// EnabledPreemptable defines whether preemptableFn is enabled
	EnabledPreemptable *bool `yaml:"enablePreemptable"`
	// EnabledReclaimable defines whether reclaimableFn is enabled
	EnabledReclaimable *bool `yaml:"enableReclaimable"`
	// EnabledQueueOrder defines whether queueOrderFn is enabled
	EnabledQueueOrder *bool `yaml:"enableQueueOrder"`
	// EnabledPredicate defines whether predicateFn is enabled
	EnabledPredicate *bool `yaml:"enablePredicate"`
	// EnabledNodeOrder defines whether NodeOrderFn is enabled
	EnabledNodeOrder *bool `yaml:"enableNodeOrder"`
	// Starvation Threshold. Jobs that have been pending longer than
	// the threshold are considered as starving jobs.
	StarvationThreshold time.Duration `yaml:"starvation-threshold"`
	// Arguments defines the different arguments that can be given to different plugins
	Arguments map[string]string `yaml:"arguments"`
}
