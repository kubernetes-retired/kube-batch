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

package schedulercache

import (
	"k8s.io/api/core/v1"

)

type Resource struct {
	MilliCPU float64
	Memory float64
	NvidiaGPU int64
}

func EmptyResource() *Resource {
	return &Resource{
		MilliCPU: 0,
		Memory: 0,
		NvidiaGPU: 0
	}
}

func NewResource(rl v1.ResourceList) {
	cpu := rl[v1.ResourceCPU]
	mem := rl[v1.ResourceMemory]
	gpu := rl[v1.ResourceNvidiaGPU]

	return &Resource{
		MilliCPU: float64(cpu.MilliValue()),
		Memory: float64(mem.Value()),
		NvidiaGPU: int64(gpu)
	}
}