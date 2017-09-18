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

package arbitrator

import (
	"testing"

	"github.com/kubernetes-incubator/kube-arbitrator/test/integration/framework"
	"k8s.io/kubernetes/pkg/api/testapi"
	"k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"github.com/kubernetes-incubator/kube-arbitrator/pkg/client"
	apiv1 "github.com/kubernetes-incubator/kube-arbitrator/pkg/apis/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestArbitrator(t *testing.T) {
	// start master
	_, s, closeFn := framework.RunAMaster(framework.NewIntegrationTestMasterConfig())
	defer closeFn()

	config := &restclient.Config{
		Host: s.URL,
		ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Groups[v1.GroupName].GroupVersion()},
	}

	// create clientset
	clientSet := clientset.NewForConfigOrDie(config)
	defer clientSet.CoreV1().Nodes().DeleteCollection(nil, metav1.ListOptions{})

	// create two namespaces
	ns01 := framework.CreateTestingNamespace("ns01", s, t)
	defer framework.DeleteTestingNamespace(ns01, s, t)
	ns02 := framework.CreateTestingNamespace("ns02", s, t)
	defer framework.DeleteTestingNamespace(ns02, s, t)
	ns01, err := clientSet.CoreV1().Namespaces().Create(ns01)
	if err != nil {
		t.Fatal("fail to create namespace ns01")
	}
	ns02, err = clientSet.CoreV1().Namespaces().Create(ns02)
	if err != nil {
		t.Fatal("fail to create namespace ns02")
	}

	// create one node
	baseNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:         "node01",
			GenerateName: "node01",
		},
		Spec: v1.NodeSpec{
			ExternalID: "foo",
		},
		Status: v1.NodeStatus{
			Capacity: v1.ResourceList{
				v1.ResourcePods:   *resource.NewQuantity(110, resource.DecimalSI),
				v1.ResourceCPU:    resource.MustParse("15"),
				v1.ResourceMemory: resource.MustParse("15Gi"),
			},
			Phase: v1.NodeRunning,
			Conditions: []v1.NodeCondition{
				{Type: v1.NodeReady, Status: v1.ConditionTrue},
			},
		},
	}
	baseNode, err = clientSet.CoreV1().Nodes().Create(baseNode)
	if err != nil {
		t.Fatalf("fail to create node baseNode, %#v\n", err)
	}

	// create two quotas
	rq01 := &v1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rq01",
			Namespace: "ns01",
		},
		Spec: v1.ResourceQuotaSpec{
			Hard: v1.ResourceList{
				v1.ResourcePods: resource.MustParse("1000"),
			},
		},
	}
	rq02 := &v1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rq02",
			Namespace: "ns02",
		},
		Spec: v1.ResourceQuotaSpec{
			Hard: v1.ResourceList{
				v1.ResourcePods: resource.MustParse("1000"),
			},
		},
	}
	rq01, err = clientSet.CoreV1().ResourceQuotas(rq01.Namespace).Create(rq01)
	if err != nil {
		t.Fatal("fail to create quota rq01")
	}
	rq02, err = clientSet.CoreV1().ResourceQuotas(rq02.Namespace).Create(rq02)
	if err != nil {
		t.Fatal("fail to create quota rq02")
	}

	// create crd and two crd samples
	extensionscs, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		t.Fatal("fail to create crd config")
	}

	_, err = client.CreateResourceQuotaAllocatorCRD(extensionscs)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatalf("fail to create crd, %#v", err)
	}
	crdClient, _, err := client.NewClient(config)
	if err != nil {
		t.Fatal("fail to create crd client")
	}
	crd01 := &apiv1.ResourceQuotaAllocator{
		ObjectMeta: metav1.ObjectMeta{
			Name: "allocator01",
			Namespace: "ns01",
		},
		Spec: apiv1.ResourceQuotaAllocatorSpec{
			Share : map[string]intstr.IntOrString{
				"weight": intstr.FromInt(1),
			},
		},
	}
	crd02 := &apiv1.ResourceQuotaAllocator{
		ObjectMeta: metav1.ObjectMeta{
			Name: "allocator02",
			Namespace: "ns02",
		},
		Spec: apiv1.ResourceQuotaAllocatorSpec{
			Share : map[string]intstr.IntOrString{
				"weight": intstr.FromInt(2),
			},
		},
	}
	var result apiv1.ResourceQuotaAllocator
	err = crdClient.Post().
		Resource(apiv1.ResourceQuotaAllocatorPlural).
		Namespace(crd01.Namespace).
		Body(crd01).
		Do().Into(&result)
	if err != nil {
		t.Fatal("fail to create crd crd01")
	}
	err = crdClient.Post().
		Resource(apiv1.ResourceQuotaAllocatorPlural).
		Namespace(crd02.Namespace).
		Body(crd02).
		Do().Into(&result)
	if err != nil {
		t.Fatal("fail to create crd crd02")
	}
}