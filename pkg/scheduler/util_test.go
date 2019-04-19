/*
Copyright 2019 The Kubernetes Authors.

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

package scheduler

import (
	"reflect"
	"testing"

	_ "github.com/kubernetes-sigs/kube-batch/pkg/scheduler/actions"
	"github.com/kubernetes-sigs/kube-batch/pkg/scheduler/conf"
)

func TestLoadSchedulerConf(t *testing.T) {
	configurations := []string{`
actions:
- name: allocate
- name: reserve 
  arguments:
    enabled: true
    threshold: 48h
tiers:
- plugins:
  - name: priority
  - name: gang
  - name: conformance
- plugins:
  - name: drf
  - name: predicates
  - name: proportion
  - name: nodeorder
`,
	}

	expectedOptions := []map[string]string{
		{
			"enabled":   "true",
			"threshold": "48h",
		},
	}

	for i, config := range configurations {
		testLoadSchedulerConfCases(config, expectedOptions[i], t)
	}
}

func testLoadSchedulerConfCases(configuration string, expectedOption map[string]string, t *testing.T) {
	trueValue := true
	expectedTiers := []conf.Tier{
		{
			Plugins: []conf.PluginOption{
				{
					Name:                "priority",
					EnabledJobOrder:     &trueValue,
					EnabledJobReady:     &trueValue,
					EnabledJobPipelined: &trueValue,
					EnabledTaskOrder:    &trueValue,
					EnabledPreemptable:  &trueValue,
					EnabledReclaimable:  &trueValue,
					EnabledQueueOrder:   &trueValue,
					EnabledPredicate:    &trueValue,
					EnabledNodeOrder:    &trueValue,
				},
				{
					Name:                "gang",
					EnabledJobOrder:     &trueValue,
					EnabledJobReady:     &trueValue,
					EnabledJobPipelined: &trueValue,
					EnabledTaskOrder:    &trueValue,
					EnabledPreemptable:  &trueValue,
					EnabledReclaimable:  &trueValue,
					EnabledQueueOrder:   &trueValue,
					EnabledPredicate:    &trueValue,
					EnabledNodeOrder:    &trueValue,
				},
				{
					Name:                "conformance",
					EnabledJobOrder:     &trueValue,
					EnabledJobReady:     &trueValue,
					EnabledJobPipelined: &trueValue,
					EnabledTaskOrder:    &trueValue,
					EnabledPreemptable:  &trueValue,
					EnabledReclaimable:  &trueValue,
					EnabledQueueOrder:   &trueValue,
					EnabledPredicate:    &trueValue,
					EnabledNodeOrder:    &trueValue,
				},
			},
		},
		{
			Plugins: []conf.PluginOption{
				{
					Name:                "drf",
					EnabledJobOrder:     &trueValue,
					EnabledJobReady:     &trueValue,
					EnabledJobPipelined: &trueValue,
					EnabledTaskOrder:    &trueValue,
					EnabledPreemptable:  &trueValue,
					EnabledReclaimable:  &trueValue,
					EnabledQueueOrder:   &trueValue,
					EnabledPredicate:    &trueValue,
					EnabledNodeOrder:    &trueValue,
				},
				{
					Name:                "predicates",
					EnabledJobOrder:     &trueValue,
					EnabledJobReady:     &trueValue,
					EnabledJobPipelined: &trueValue,
					EnabledTaskOrder:    &trueValue,
					EnabledPreemptable:  &trueValue,
					EnabledReclaimable:  &trueValue,
					EnabledQueueOrder:   &trueValue,
					EnabledPredicate:    &trueValue,
					EnabledNodeOrder:    &trueValue,
				},
				{
					Name:                "proportion",
					EnabledJobOrder:     &trueValue,
					EnabledJobReady:     &trueValue,
					EnabledJobPipelined: &trueValue,
					EnabledTaskOrder:    &trueValue,
					EnabledPreemptable:  &trueValue,
					EnabledReclaimable:  &trueValue,
					EnabledQueueOrder:   &trueValue,
					EnabledPredicate:    &trueValue,
					EnabledNodeOrder:    &trueValue,
				},
				{
					Name:                "nodeorder",
					EnabledJobOrder:     &trueValue,
					EnabledJobReady:     &trueValue,
					EnabledJobPipelined: &trueValue,
					EnabledTaskOrder:    &trueValue,
					EnabledPreemptable:  &trueValue,
					EnabledReclaimable:  &trueValue,
					EnabledQueueOrder:   &trueValue,
					EnabledPredicate:    &trueValue,
					EnabledNodeOrder:    &trueValue,
				},
			},
		},
	}

	config, err := loadSchedulerConf(configuration)
	if err != nil {
		t.Errorf("Failed to load scheduler configuration: %v", err)
	}
	tiers := config.Tiers
	if !reflect.DeepEqual(tiers, expectedTiers) {
		t.Errorf("Failed to set default settings for plugins, expected: %+v, got %+v",
			expectedTiers, tiers)
	}

	if len(config.Actions) != 2 ||
		config.Actions[0].Name != "allocate" ||
		len(config.Actions[0].Arguments) != 0 {
		t.Errorf("Failed to set default settings for backfill, "+
			"expected: len(config.Actions)==2, config.Action[1].Name=='allocate', config.Action[1].Options==[], got %+v",
			config.Actions)
	}

	if config.Actions[1].Name != "reserve" ||
		len(config.Actions[1].Arguments) != 2 ||
		config.Actions[1].Arguments["enabled"] != expectedOption["enabled"] ||
		config.Actions[1].Arguments["threshold"] != expectedOption["threshold"] {
		t.Errorf("Failed to set default settings for reserve, "+
			"expected: len(config.Actions)==2, config.Action[1].Name=='reserve', "+
			"config.Action[1].Options[\"enabled\"]==\"%s\", got %+v",
			expectedOption, config.Actions)
	}

}
