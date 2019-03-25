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

package webhook

import (
	"net/http"

	"github.com/golang/glog"
	"k8s.io/client-go/rest"

	kbver "github.com/kubernetes-sigs/kube-batch/pkg/client/clientset/versioned"
	"github.com/kubernetes-sigs/kube-batch/pkg/webhook/api"
	"github.com/kubernetes-sigs/kube-batch/pkg/webhook/util"
	"github.com/kubernetes-sigs/kube-batch/pkg/webhook/validating"
)

// Server will start a https server for admitting.
type Server struct {
	admitterContext api.AdmitterContext

	listenAddress string
	certFile      string
	keyFile       string
}

// NewServer create a new Server for admitting.
func NewServer(config *rest.Config, listenAddress, certFile, keyFile string) (*Server, error) {
	server := &Server{
		listenAddress: listenAddress,
		certFile:      certFile,
		keyFile:       keyFile,
	}

	kubeclient := kbver.NewForConfigOrDie(config)

	server.admitterContext = api.AdmitterContext{
		KubeClient: kubeclient,
	}
	return server, nil
}

// Run starts informers, and listens for accepting request.
func (ws *Server) Run(stopCh <-chan struct{}) {
	mux := http.NewServeMux()
	if pgv, err := validating.NewPodGroupValidator(ws.admitterContext); err != nil {
		panic(err)
	} else {
		mux.HandleFunc("/validate/podgroup", func(w http.ResponseWriter, r *http.Request) {
			util.Serve(w, r, pgv)
		})
	}
	server := &http.Server{
		Addr:    ws.listenAddress,
		Handler: mux,
	}
	glog.Fatal(server.ListenAndServeTLS(ws.certFile, ws.keyFile))
}
