#!/bin/bash

export ROOT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/..
export LOG_LEVEL=3
export CLEANUP_CLUSTER=${CLEANUP_CLUSTER:-1}
export CLUSTER_CONTEXT="--name test"
export IMAGE_NGINX="nginx:latest"
export IMAGE_BUSYBOX="busybox:latest"
export KIND_OPT=${KIND_OPT:=" --config ${ROOT_DIR}/hack/e2e-kind-config.yaml"}
export KA_BIN=_output/bin
export WAIT_TIME="20s"


sudo apt-get update && sudo apt-get install -y apt-transport-https
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee -a /etc/apt/sources.list.d/kubernetes.list
sudo apt-get update
sudo apt-get install -y kubectl

# Download kind binary (0.2.0)
sudo curl -o /usr/local/bin/kind -L https://github.com/kubernetes-sigs/kind/releases/download/0.2.0/kind-linux-amd64
sudo chmod +x /usr/local/bin/kind

# check if kind installed
function check-prerequisites {
  echo "checking prerequisites"
  which kind >/dev/null 2>&1
  if [[ $? -ne 0 ]]; then
    echo "kind not installed, exiting."
    exit 1
  else
    echo -n "found kind, version: " && kind version
  fi

  which kubectl >/dev/null 2>&1
  if [[ $? -ne 0 ]]; then
    echo "kubectl not installed, exiting."
    exit 1
  else
    echo -n "found kubectl, " && kubectl version --short --client
  fi
}

function kind-up-cluster {
  check-prerequisites
  echo "Running kind: [kind create cluster ${CLUSTER_CONTEXT} ${KIND_OPT}]"
  kind create cluster ${CLUSTER_CONTEXT} ${KIND_OPT} --wait ${WAIT_TIME}
  docker pull ${IMAGE_BUSYBOX}
  docker pull ${IMAGE_NGINX}
  kind load docker-image ${IMAGE_NGINX} ${CLUSTER_CONTEXT}
  kind load docker-image ${IMAGE_BUSYBOX} ${CLUSTER_CONTEXT}
}

# clean up
function cleanup {
    killall -9 kube-batch
    kind delete cluster ${CLUSTER_CONTEXT}

    echo "===================================================================================="
    echo "=============================>>>>> Scheduler Logs <<<<<============================="
    echo "===================================================================================="

    cat scheduler.log
}

function kube-batch-up {
    cd ${ROOT_DIR}

    export KUBECONFIG="$(kind get kubeconfig-path ${CLUSTER_CONTEXT})"
    kubectl version

    kubectl create -f deployment/kube-batch/templates/scheduling_v1alpha1_queue.yaml
    kubectl create -f deployment/kube-batch/templates/scheduling_v1alpha1_podgroup.yaml
    kubectl create -f deployment/kube-batch/templates/scheduling_v1alpha2_podgroup.yaml
    kubectl create -f deployment/kube-batch/templates/scheduling_v1alpha2_queue.yaml
    kubectl create -f deployment/kube-batch/templates/default.yaml

    # start kube-batch
    nohup ${KA_BIN}/kube-batch --kubeconfig ${KUBECONFIG} --scheduler-conf=config/kube-batch-conf.yaml --logtostderr --v ${LOG_LEVEL} > scheduler.log 2>&1 &
}


trap cleanup EXIT

kind-up-cluster

kube-batch-up

cd ${ROOT_DIR}
go test ./test/e2e -v -timeout 30m
