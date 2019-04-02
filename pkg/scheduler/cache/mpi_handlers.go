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

package cache

import (
	"github.com/golang/glog"

	"k8s.io/client-go/tools/cache"

	kbv1 "github.com/kubernetes-sigs/kube-batch/pkg/apis/scheduling/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime/schema"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const kubeexecsh = `#!/bin/sh
echo "kubectl executor starting for $1"
set -x
POD_NAME=$1
shift
/usr/bin/kubectl exec ${POD_NAME} -- /bin/sh -c "$*"
/usr/bin/kubectl exec ${POD_NAME} -- /bin/sh -c "echo $? > /tmp/exit"`

const startupsh = `#!/bin/sh
while true; do
	cat /entry/hosts > /etc/hosts
	if [ "$(cat /entry/executor)" = "$(hostname)" ]; then
		echo "Look at me. I'm the executor now.";
		break
	fi
	if [ -f /tmp/exit ]; then
		exit $(cat /tmp/exit);
	fi
	sleep 1
done
wget -nv https://storage.googleapis.com/kubernetes-release/release/v1.13.4/bin/linux/amd64/kubectl && \
mv kubectl /usr/bin/kubectl && \
chmod +x /usr/bin/kubectl

while true; do
	happy=0;
	while read p; do
		/usr/bin/kubectl exec ${p} -- /bin/sh -c echo "My name is $(hostname)";
		happy=$((happy+$?));
	done < /mpi/hostfile;

	if [ $happy = 0 ]; then
		echo "We're happy -- $happy";
		break;
	fi
	sleep 3;
done

echo "Time to do the thing: $*"
$*`

func (sc *SchedulerCache) AddMPI(obj interface{}) {
	ss, ok := obj.(*kbv1.MPI)
	if !ok {
		glog.Errorf("Cannot convert to *kbv1.MPI: %v", obj)
		return
	}

	sc.Mutex.Lock()
	defer sc.Mutex.Unlock()

	glog.V(4).Infof("Add MPI(%s) into cache, spec(%#v)", ss.Name, ss.Spec)
	err := sc.addMPI(ss)
	if err != nil {
		glog.Errorf("Failed to add MPI %s into cache: %v", ss.Name, err)
		return
	}
	return
}

func (sc *SchedulerCache) UpdateMPI(oldObj, newObj interface{}) {
	oldSS, ok := oldObj.(*kbv1.MPI)
	if !ok {
		glog.Errorf("Cannot convert oldObj to *kbv1.MPI: %v", oldObj)
		return
	}
	newSS, ok := newObj.(*kbv1.MPI)
	if !ok {
		glog.Errorf("Cannot convert newObj to *kbv1.MPI: %v", newObj)
		return
	}

	sc.Mutex.Lock()
	defer sc.Mutex.Unlock()

	err := sc.updateMPI(oldSS, newSS)
	if err != nil {
		glog.Errorf("Failed to update MPI %s into cache: %v", oldSS.Name, err)
		return
	}
	return
}

func (sc *SchedulerCache) DeleteMPI(obj interface{}) {
	var ss *kbv1.MPI
	switch t := obj.(type) {
	case *kbv1.MPI:
		ss = t
	case cache.DeletedFinalStateUnknown:
		var ok bool
		ss, ok = t.Obj.(*kbv1.MPI)
		if !ok {
			glog.Errorf("Cannot convert to *kbv1.MPI: %v", t.Obj)
			return
		}
	default:
		glog.Errorf("Cannot convert to *kbv1.MPI: %v", t)
		return
	}

	sc.Mutex.Lock()
	defer sc.Mutex.Unlock()

	err := sc.deleteMPI(ss)
	if err != nil {
		glog.Errorf("Failed to delete MPI %s from cache: %v", ss.Name, err)
		return
	}
	return
}

func (sc *SchedulerCache) AddMPIPod(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		glog.Errorf("Cannot convert to *kbv1.MPI: %v", obj)
		return
	}

	sc.Mutex.Lock()
	defer sc.Mutex.Unlock()

	err := sc.updateMPIConfigmap(pod)
	if err != nil {
		glog.Errorf("Failed to update MPI Pod Configmap %s in cache: %v", pod.Name, err)
	}

	return
}

func (sc *SchedulerCache) UpdateMPIPod(oldObj, newObj interface{}) {
	oldPod, ok := oldObj.(*corev1.Pod)
	if !ok {
		glog.Errorf("Cannot convert oldObj to *kbv1.MPI: %v", oldObj)
		return
	}
	newPod, ok := newObj.(*corev1.Pod)
	if !ok {
		glog.Errorf("Cannot convert newObj to *kbv1.MPI: %v", newObj)
		return
	}

	glog.Infof("Updating MPI Pod from and to: %v %v", oldPod, newPod)

	return
}

func (sc *SchedulerCache) DeleteMPIPod(obj interface{}) {
	ss, ok := obj.(*corev1.Pod)
	if !ok {
		glog.Errorf("Cannot convert to *kbv1.MPI: %v", obj)
		return
	}

	glog.Infof("Delete MPI Pod: %v", ss)
	return
}

func (sc *SchedulerCache) addMPI(mpi *kbv1.MPI) error {
	// Here we create all the children objects to the MPI object
	glog.Infof("New MPI thing: %v", mpi.GetName())

	gvk := schema.GroupVersionKind{
		Group:   kbv1.GroupName,
		Version: kbv1.GroupVersion,
		Kind:    "MPI",
	}

	sa := generateServiceAccount(mpi, gvk)
	role := generateRole(mpi, gvk)
	roleBinding := generateRoleBinding(mpi, gvk)

	sc.kubeclient.CoreV1().ServiceAccounts(mpi.GetNamespace()).Create(sa)
	sc.kubeclient.RbacV1().Roles(mpi.GetNamespace()).Create(role)
	sc.kubeclient.RbacV1().RoleBindings(mpi.GetNamespace()).Create(roleBinding)

	// Generate the mpi-hostfile placeholder
	sc.kubeclient.CoreV1().ConfigMaps(mpi.GetNamespace()).Create(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mpi.GetName() + "-hostfile",
			Namespace: mpi.GetNamespace(),
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mpi, gvk),
			},
		},
		Data: map[string]string{
			"hostfile": "",
		},
	})
	// Generate the mpi-data configmap
	sc.kubeclient.CoreV1().ConfigMaps(mpi.GetNamespace()).Create(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mpi.GetName() + "-mpi-data",
			Namespace: mpi.GetNamespace(),
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mpi, gvk),
			},
		},
		Data: map[string]string{
			"executor":    "",
			"hosts":       "",
			"kubeexec.sh": kubeexecsh,
			"startup.sh":  startupsh,
		},
	})

	job := GenerateJob(mpi, mpi.Spec.Job)

	sc.kubeclient.BatchV1().Jobs(mpi.GetNamespace()).Create(job)

	// Define the desired PodGroup object
	podgroup := &kbv1.PodGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mpi.GetName() + "-podgroup",
			Namespace: mpi.GetNamespace(),
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mpi, gvk),
			},
		},
		Spec: kbv1.PodGroupSpec{
			MinMember: *mpi.Spec.Job.Parallelism,
		},
	}

	sc.kbclient.SchedulingV1alpha1().PodGroups(mpi.GetNamespace()).Create(podgroup)

	return nil
}

func (sc *SchedulerCache) updateMPI(oldObj, newObj *kbv1.MPI) error {
	sc.deleteMPI(oldObj)
	sc.addMPI(newObj)

	return nil
}

func (sc *SchedulerCache) deleteMPI(mpi *kbv1.MPI) error {
	glog.Infof("Delete MPI thing: %v", mpi.GetName())

	return nil
}

func (sc *SchedulerCache) updateMPIConfigmap(pod *corev1.Pod) error {
	glog.Infof("Update MPI Configmap for : %v", pod.GetName())
	pods, err := sc.kubeclient.CoreV1().Pods(pod.GetNamespace()).List(metav1.ListOptions{
		LabelSelector: "job-name=" + pod.GetLabels()["job-name"],
	})
	if err != nil {
		glog.Errorf("Error getting a pod list for mpi-job %v\n", pod.GetLabels()["job-name"])
	}
	ipMap := make(map[string]string)
	hostMap := []string{}
	hostfileString := ""
	etcHostsString := ""
	for _, v := range pods.Items {
		if v.Status.PodIP != "" {
			ipMap[v.GetName()] = v.Status.PodIP
			hostMap = append(hostMap, v.GetName())
			hostfileString = hostfileString + v.GetName() + "\n"
			etcHostsString = etcHostsString + v.Status.PodIP + " " + v.GetName() + "\n"
		}
	}
	if len(ipMap) == len(pods.Items) {

		hostfileMapName := pod.GetLabels()["job-name"] + "-hostfile"
		dataMapName := pod.GetLabels()["job-name"] + "-mpi-data"
		glog.Infof("We have IPs, time to update %v\n", hostfileMapName)
		glog.Infof("%v\n", ipMap)

		gvk := schema.GroupVersionKind{
			Group:   kbv1.GroupName,
			Version: kbv1.GroupVersion,
			Kind:    "MPI",
		}

		mpi, err := sc.kbclient.SchedulingV1alpha1().MPIs(pod.GetNamespace()).Get(pod.GetLabels()["job-name"], metav1.GetOptions{})
		if err != nil {
			return err
		}

		sc.kubeclient.CoreV1().ConfigMaps(pod.GetNamespace()).Update(
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      hostfileMapName,
					Namespace: pod.GetNamespace(),
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(mpi, gvk),
					},
				},
				Data: map[string]string{
					"hostfile": hostfileString,
				},
			},
		)
		glog.Infof("Our executor is now %v\n", pods.Items[0].Name)
		sc.kubeclient.CoreV1().ConfigMaps(pod.GetNamespace()).Update(
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      dataMapName,
					Namespace: pod.GetNamespace(),
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(pod, gvk),
					},
				},
				Data: map[string]string{
					"executor":    pods.Items[0].Name,
					"kubeexec.sh": kubeexecsh,
					"startup.sh":  startupsh,
					"hosts":       etcHostsString,
				},
			},
		)

		role, err := sc.kubeclient.RbacV1().Roles(pod.GetNamespace()).Get(pod.GetLabels()["job-name"]+"-role", metav1.GetOptions{})
		if err != nil {
			glog.Errorf("Error getting role %s", pod.GetLabels()["job-name"]+"-role")
		}
		for k := range role.Rules {
			role.Rules[k].ResourceNames = hostMap
		}
		sc.kubeclient.RbacV1().Roles(pod.GetNamespace()).Update(role)
	}
	return nil
}

func generateRole(mpi metav1.Object, gvk schema.GroupVersionKind) *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mpi.GetName() + "-role",
			Namespace: mpi.GetNamespace(),
			Labels: map[string]string{
				"job-name": mpi.GetName(),
			},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mpi, gvk),
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:         []string{"get"},
				APIGroups:     []string{""},
				Resources:     []string{"pods"},
				ResourceNames: []string{""},
			},
			{
				Verbs:         []string{"create"},
				APIGroups:     []string{""},
				Resources:     []string{"pods/exec"},
				ResourceNames: []string{""},
			},
		},
	}
}
func generateRoleBinding(mpi metav1.Object, gvk schema.GroupVersionKind) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mpi.GetName() + "-rolebinding",
			Namespace: mpi.GetNamespace(),
			Labels: map[string]string{
				"job-name": mpi.GetName(),
			},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mpi, gvk),
			},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      mpi.GetName() + "-sa",
				Namespace: mpi.GetNamespace(),
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     mpi.GetName() + "-role",
		},
	}
}

func generateServiceAccount(mpi metav1.Object, gvk schema.GroupVersionKind) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mpi.GetName() + "-sa",
			Namespace: mpi.GetNamespace(),
			Labels: map[string]string{
				"job-name": mpi.GetName(),
			},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mpi, gvk),
			},
		},
	}
}
