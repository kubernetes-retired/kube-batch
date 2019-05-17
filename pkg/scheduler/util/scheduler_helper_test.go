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

package util

import (
	"reflect"
	"testing"
)

func TestSelectBestNode(t *testing.T) {
	cases := []struct {
		PriorityList HostPriorityList
		// Expected node is one of ExpectedNodes
		ExpectedNodes []string
	}{
		{
			PriorityList: HostPriorityList{
				{
					Host:  "node1",
					Score: 1.0,
				},
				{
					Host:  "node2",
					Score: 1.0,
				},
				{
					Host:  "node3",
					Score: 2.0,
				},
				{
					Host:  "node4",
					Score: 2.0,
				},
			},
			ExpectedNodes: []string{"node3", "node4"},
		},
		{
			PriorityList: HostPriorityList{
				{
					Host:  "node1",
					Score: 1.0,
				},
				{
					Host:  "node2",
					Score: 1.0,
				},
				{
					Host:  "node3",
					Score: 3.0,
				},
				{
					Host:  "node4",
					Score: 2.0,
				},
				{
					Host:  "node5",
					Score: 2.0,
				},
			},
			ExpectedNodes: []string{"node3"},
		},
	}

	oneOf := func(node string, nodes []string) bool {
		for _, v := range nodes {
			if reflect.DeepEqual(node, v) {
				return true
			}
		}
		return false
	}
	for i, test := range cases {
		result := SelectBestNode(test.PriorityList)
		if !oneOf(result, test.ExpectedNodes) {
			t.Errorf("Failed test case #%d, expected: %#v, got %#v", i, test.ExpectedNodes, result)
		}
	}
}
