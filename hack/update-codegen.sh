#!/usr/bin/env bash

go get -u k8s.io/code-generator

cd $GOPATH/src/k8s.io/code-generator

./generate-groups.sh all "github.com/kubernetes-sigs/kube-batch/pkg/client" "github.com/kubernetes-sigs/kube-batch/pkg/apis" scheduling:v1alpha1

cd $GOPATH/src/github.com/kubernetes-sigs/kube-batch/