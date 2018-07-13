BIN_DIR=_output/bin
DOCKER_REPO=damora
RELEASE_VER=0.1

kube-arbitrator: init
	go build -o ${BIN_DIR}/kar-scheduler ./cmd/kar-scheduler/
	go build -o ${BIN_DIR}/kar-controllers ./cmd/kar-controllers/
	go build -o ${BIN_DIR}/karcli ./cmd/karcli

verify: generate-code
	hack/verify-gofmt.sh
	hack/verify-golint.sh
	hack/verify-gencode.sh

init:
	mkdir -p ${BIN_DIR}

generate-code:
	go build -o ${BIN_DIR}/deepcopy-gen ./cmd/deepcopy-gen/
	${BIN_DIR}/deepcopy-gen -i ./pkg/apis/v1alpha1/ -O zz_generated.deepcopy

images: kube-arbitrator
	cp ./_output/bin/kar-scheduler ./deployment/
	cp ./_output/bin/kar-controllers ./deployment/
	cp ./_output/bin/karcli ./deployment/
	docker build --no-cache -f ./deployment/Dockerfile.ppc64le.ubuntu ./deployment/  \
	-t ${DOCKER_REPO}/ppc64le-kube-arbitrator:ubuntu-${RELEASE_VER}
	docker build --no-cache -f ./deployment/Dockerfile.ppc64le.centos ./deployment/  \
	-t ${DOCKER_REPO}/ppc64le-kube-arbitrator:centos-${RELEASE_VER}
	docker build --no-cache -f ./deployment/Dockerfile.amd64.ubuntu ./deployment/  \
	-t ${DOCKER_REPO}/amd64-kube-arbitrator:ubuntu-${RELEASE_VER} 
	rm -f ./deployment/kar*

manifest:
	docker push ${DOCKER_REPO}/ppc64le-kube-arbitrator:ubuntu-${RELEASE_VER}
	docker push ${DOCKER_REPO}/ppc64le-kube-arbitrator:centos-${RELEASE_VER}
	docker push ${DOCKER_REPO}/amd64-kube-arbitrator:ubuntu-${RELEASE_VER} 
	docker manifest create -a ${DOCKER_REPO}/kube-arbitrator:${RELEASE_VER}  	\
	${DOCKER_REPO}/amd64-kube-arbitrator:ubuntu-${RELEASE_VER}			\
	${DOCKER_REPO}/ppc64le-kube-arbitrator:centos-${RELEASE_VER}  			\
	${DOCKER_REPO}/ppc64le-kube-arbitrator:ubuntu-${RELEASE_VER}
	docker manifest annotate ${DOCKER_REPO}/kube-arbitrator:${RELEASE_VER} 		\
	${DOCKER_REPO}/amd64-kube-arbitrator:ubuntu-${RELEASE_VER} --arch amd64
	docker manifest annotate  ${DOCKER_REPO}/kube-arbitrator:${RELEASE_VER} 	\
	${DOCKER_REPO}/ppc64le-kube-arbitrator:centos-${RELEASE_VER}  --arch ppc64le 
	docker manifest annotate ${DOCKER_REPO}/kube-arbitrator:${RELEASE_VER} 		\
	${DOCKER_REPO}/ppc64le-kube-arbitrator:ubuntu-${RELEASE_VER}  --arch ppc64le
	docker manifest push  ${DOCKER_REPO}/kube-arbitrator:${RELEASE_VER}

run-test:
	hack/make-rules/test.sh $(WHAT) $(TESTS)

clean:
	rm -rf _output/
	rm -f kube-arbitrator
