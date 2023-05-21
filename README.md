# kube-batch


[![Build Status](https://travis-ci.org/kubernetes-sigs/kube-batch.svg?branch=master)](https://travis-ci.org/kubernetes-sigs/kube-batch)
[![Go Report Card](https://goreportcard.com/badge/github.com/kubernetes-sigs/kube-batch)](https://goreportcard.com/report/github.com/kubernetes-sigs/kube-batch)
[![RepoSize](https://img.shields.io/github/repo-size/kubernetes-sigs/kube-batch.svg)](https://github.com/kubernetes-sigs/kube-batch)
[![Release](https://img.shields.io/github/release/kubernetes-sigs/kube-batch.svg)](https://github.com/kubernetes-sigs/kube-batch/releases)
[![LICENSE](https://img.shields.io/github/license/kubernetes-sigs/kube-batch.svg)](https://github.com/kubernetes-sigs/kube-batch/blob/master/LICENSE)

`kube-batch` is a batch scheduler for Kubernetes, providing mechanisms for applications which would like to run batch jobs leveraging Kubernetes. It builds upon a decade and a half of experience on running batch workloads at scale using several systems, combined with best-of-breed ideas and practices from the open source community.

Refer to [tutorial](doc/usage/tutorial.md) on how to use `kube-batch` to run batch job in Kubernetes.

## Overall Architecture

The following figure describes the overall architecture and scope of `kube-batch`; the out-of-scope part is going to be handled by other projects.

![kube-batch](doc/images/kube-batch.png)

## Who uses kube-batch?

As the kube-batch Community grows, we'd like to keep track of our users. Please send a PR with your organization name.

Currently **officially** using kube-batch:

1. [openBCE](https://github.com/openbce)
1. [Kubeflow](https://www.kubeflow.org)
1. [Volcano](https://github.com/volcano-sh/volcano)
1. [Baidu Inc](http://www.baidu.com)
1. [TuSimple](https://www.tusimple.com)
1. [MOGU Inc](https://www.mogujie.com)
1. [Vivo](https://www.vivo.com)

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- Slack: #sig-scheduling
- Mailing List: https://groups.google.com/forum/#!forum/kubernetes-sig-scheduling

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
