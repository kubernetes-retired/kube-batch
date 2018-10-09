#!/bin/bash

export dest_dir="`pwd`/_output/bin"
export top_dir=`pwd`

function download_bins() {
	quiet="-s"
	BASE_URL="https://storage.googleapis.com/k8s-c10s-test-binaries"
	
	os="$(uname -s)"
	os_lowercase="$(echo "$os" | tr '[:upper:]' '[:lower:]' )"
	arch="$(uname -m)"
	
	etcd_dest="${dest_dir}/etcd"
	kube_apiserver_dest="${dest_dir}/kube-apiserver"
	kubectl_dest="${dest_dir}/kubectl"
	
	if [ ! -f "$etcd_dest" ]; then
		echo "Downloading ETCD ......"
		curl $quiet "${BASE_URL}/etcd-${os}-${arch}" --output "$etcd_dest"
		chmod +x $etcd_dest
	fi

	if [ ! -f "$kube_apiserver_dest" ]; then
		echo "Downloading kube-apiserver ......"
		curl $quiet "${BASE_URL}/kube-apiserver-${os}-${arch}" --output "$kube_apiserver_dest"
		chmod +x $kube_apiserver_dest
	fi


	if [ ! -f "$kubectl_dest" ]; then
		echo "Downloading kubectl ......"
	
		kubectl_version="$(curl $quiet https://storage.googleapis.com/kubernetes-release/release/stable.txt)"
		kubectl_url="https://storage.googleapis.com/kubernetes-release/release/${kubectl_version}/bin/${os_lowercase}/amd64/kubectl"
	
		curl $quiet "$kubectl_url" --output "$kubectl_dest"
	
		chmod +x $kubectl_dest
	fi
}

download_bins

echo    "# destination:"
echo    "#   ${dest_dir}"
echo    "# versions:"
echo -n "#   etcd:            "; "$etcd_dest" --version | head -n 1
echo -n "#   kube-apiserver:  "; "$kube_apiserver_dest" --version
echo -n "#   kubectl:         "; "$kubectl_dest" version --client --short

echo ""

go test ./test/integration -v
