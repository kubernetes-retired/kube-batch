package controller

import (
	qjobv1 "github.com/kubernetes-incubator/kube-arbitrator/pkg/apis/v1"
)

// GetPodFullName returns a name that uniquely identifies a qj.
func GetQJFullName(qj *qjobv1.QueueJob) string {
	// Use underscore as the delimiter because it is not allowed in qj name
	// (DNS subdomain format).
	return qj.Name + "_" + qj.Namespace
}

func HigherPriorityQJ(qj1, qj2 interface{} ) bool {
	return (qj1.(*qjobv1.QueueJob).Spec.Priority > qj2.(*qjobv1.QueueJob).Spec.Priority)
}