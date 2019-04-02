BIN_DIR=_output/bin
RELEASE_VER=v0.4.2
REPO_PATH=github.com/kubernetes-sigs/kube-batch
GitSHA=`git rev-parse HEAD`
Date=`date "+%Y-%m-%d %H:%M:%S"`
REL_OSARCH="linux/amd64"
LD_FLAGS=" \
    -X '${REPO_PATH}/pkg/version.GitSHA=${GitSHA}' \
    -X '${REPO_PATH}/pkg/version.Built=${Date}'   \
    -X '${REPO_PATH}/pkg/version.Version=${RELEASE_VER}'"

kube-batch: init
	go build -ldflags ${LD_FLAGS} -o=${BIN_DIR}/kube-batch ./cmd/kube-batch

verify: generate-code
	hack/verify-gofmt.sh
	hack/verify-golint.sh
	hack/verify-gencode.sh

init:
	mkdir -p ${BIN_DIR}

generate-code:
	go get -u k8s.io/code-generator/...
	hack/update-codegen.sh

rel_bins:
	go get github.com/mitchellh/gox
	CGO_ENABLED=0 gox -osarch=${REL_OSARCH} -ldflags ${LD_FLAGS} \
	-output=${BIN_DIR}/{{.OS}}/{{.Arch}}/kube-batch ./cmd/kube-batch

images: rel_bins
	cp ./_output/bin/${REL_OSARCH}/kube-batch ./deployment/images/
	docker build ./deployment/images -t kubesigs/kube-batch:${RELEASE_VER}
	rm -f ./deployment/images/kube-batch

run-test:
	hack/make-rules/test.sh $(WHAT) $(TESTS)

e2e: kube-batch
	hack/run-e2e.sh

coverage:
	KUBE_COVER=y hack/make-rules/test.sh $(WHAT) $(TESTS)

clean:
	rm -rf _output/
	rm -f kube-batch
