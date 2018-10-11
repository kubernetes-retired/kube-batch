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

package integration

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang/glog"
	. "github.com/onsi/gomega"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"sigs.k8s.io/testing_frameworks/integration"

	"github.com/kubernetes-sigs/kube-batch/pkg/client/clientset/versioned"
	"github.com/kubernetes-sigs/kube-batch/pkg/scheduler"
)

func startCluster(ctx *context) (*integration.ControlPlane, *kubernetes.Clientset, *versioned.Clientset) {
	destDir := os.Getenv("dest_dir")
	topDir := os.Getenv("top_dir")

	// Start master.
	apiServ := &integration.APIServer{
		Path: strings.Join([]string{destDir, "kube-apiserver"}, "/"),
	}

	etcd := &integration.Etcd{
		Path: strings.Join([]string{destDir, "etcd"}, "/"),
	}

	cp := &integration.ControlPlane{
		APIServer: apiServ,
		Etcd:      etcd,
	}
	err := cp.Start()
	Expect(err).NotTo(HaveOccurred())

	// Init cluster client.
	config, err := clientcmd.BuildConfigFromFlags(cp.APIURL().Host, "")
	Expect(err).NotTo(HaveOccurred())

	sched, err := scheduler.NewScheduler(config, "kube-batch", "", true)
	Expect(err).NotTo(HaveOccurred())
	go sched.Run(ctx.stopEverything)

	kubeclient := kubernetes.NewForConfigOrDie(config)
	karclient := versioned.NewForConfigOrDie(config)

	// Start node workers.
	createNodes(kubeclient, 5)

	go kubeletSimulator(kubeclient, ctx.stopEverything)

	// Register CRDs.
	kc := cp.KubeCtl()
	kc.Path = strings.Join([]string{destDir, "kubectl"}, "/")

	_, _, err = kc.Run("create", "-f",
		strings.Join([]string{topDir, "config/crds/scheduling_v1alpha1_queue.yaml"}, "/"))
	Expect(err).NotTo(HaveOccurred())
	_, _, err = kc.Run("create", "-f",
		strings.Join([]string{topDir, "config/crds/scheduling_v1alpha1_podgroup.yaml"}, "/"))
	Expect(err).NotTo(HaveOccurred())

	return cp, kubeclient, karclient
}

func shutdownCluster(ctx *context) {
	close(ctx.stopEverything)
	ctx.ctrlPlane.Stop()
}

func newNode(nn string) *v1.Node {
	return &v1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{{Type: v1.NodeReady, Status: v1.ConditionTrue}},
			Allocatable: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("4000m"),
				v1.ResourceMemory: resource.MustParse("8Gi"),
				v1.ResourcePods:   resource.MustParse("100"),
			},
		},
	}
}

func createNodes(kubeclient *kubernetes.Clientset, nodeNum int) {
	for i := 0; i < nodeNum; i++ {
		node := newNode(fmt.Sprintf("node-%d", i))
		if _, err := kubeclient.CoreV1().Nodes().Create(node); err != nil {
			glog.Errorf("Failed to create node %s", node.Name)
		}
	}
}

func kubeletSimulator(kubeclient *kubernetes.Clientset, stop chan struct{}) {
	for {
		select {
		case <-stop:
			return
		default:
			pods, err := kubeclient.CoreV1().Pods(v1.NamespaceAll).List(metav1.ListOptions{})
			if err != nil {
				continue
			}
			for _, pod := range pods.Items {
				switch pod.Status.Phase {
				case v1.PodPending:
					// If pod is pending && nodeName not empty, set it to running
					if pod.Spec.NodeName != "" {
						pod.Status.Phase = v1.PodRunning
						now := metav1.Now()
						pod.Status.StartTime = &now
						kubeclient.CoreV1().Pods(pod.Namespace).UpdateStatus(&pod)
					}
				case v1.PodRunning:
					if t, found := pod.Annotations[pod_exec_time_key]; found {
						now := time.Now()
						d, err := time.ParseDuration(t)
						if err != nil {
							glog.Errorf("Failed to parse pod execution time <%s>: %v", t, err)
							break
						}
						st := pod.Status.StartTime.Time.Add(d)
						if now.After(st) {
							pod.Status.Phase = v1.PodSucceeded
							kubeclient.CoreV1().Pods(pod.Namespace).UpdateStatus(&pod)
						}
					}
				}
			}
		}
	}
}
