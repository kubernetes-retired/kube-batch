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

import (
	"fmt"
	"math"
)

var ResourceNames = []string{
	"cpu", "memory", "gpu",
}

type Resource struct {
	MilliCPU float64
	Memory   float64
	GPU      float64
}

func EmptyResource() *Resource {
	return &Resource{
		MilliCPU: 0,
		Memory:   0,
		GPU:      0,
	}
}

func (r *Resource) Get(rn string) float64 {
	switch rn {
	case "cpu":
		return r.MilliCPU
	case "memory":
		return r.Memory
	case "gpu":
		return r.GPU
	}

	return 0
}

func (r *Resource) Clone() *Resource {
	clone := &Resource{
		MilliCPU: r.MilliCPU,
		Memory:   r.Memory,
		GPU:      r.GPU,
	}
	return clone
}

var minMilliCPU float64 = 10
var minMemory float64 = 10 * 1024 * 1024

func (r *Resource) IsEmpty() bool {
	return r.MilliCPU < minMilliCPU && r.Memory < minMemory && r.GPU == 0
}

func (r *Resource) Add(rr *Resource) *Resource {
	r.MilliCPU += rr.MilliCPU
	r.Memory += rr.Memory
	r.GPU += rr.GPU
	return r
}

//A function to Subtract two Resource objects.
func (r *Resource) Sub(rr *Resource) *Resource {
	if r.Less(rr) == false {
		r.MilliCPU -= rr.MilliCPU
		r.Memory -= rr.Memory
		r.GPU -= rr.GPU
		return r
	}
	panic("Resource is not sufficient to do operation: Sub()")
}

func (r *Resource) Less(rr *Resource) bool {
	return r.MilliCPU < rr.MilliCPU && r.Memory < rr.Memory && r.GPU < rr.GPU
}

func (r *Resource) LessEqual(rr *Resource) bool {
	return (r.MilliCPU < rr.MilliCPU || math.Abs(rr.MilliCPU-r.MilliCPU) < 0.01) &&
		(r.Memory < rr.Memory || math.Abs(rr.Memory-r.Memory) < 1) &&
		(r.GPU <= rr.GPU)
}

func (r *Resource) String() string {
	return fmt.Sprintf("cpu %0.2f, memory %0.2f, GPU %d",
		r.MilliCPU, r.Memory, r.GPU)
}
