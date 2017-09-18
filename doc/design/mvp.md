# Kube Arbitrator Roadmap/MVP

@k82cn, @foxish, 08/27/2017


## Overview

[Resource sharing architecture for batch and serving workloads in Kubernetes](https://docs.google.com/document/d/1-H2hnZap7gQivcSU-9j4ZrJ8wE_WwcfOkTeAGjzUyLA/edit#heading=h.a1k69dgabg0w) proposed `QueueJob` feature to run batch job with services workload in Kuberentes. Considering the complexity, the whole batch job proposal was separated into two phase: `Queue` and `QueueJob`. Refer to the following roadmap for the detail of `Queue` and `QueueJob`

## infra (MVP)

### code generated

1. Generate Deepcopy: **Done**

### test-infra

kube-arbitrator focus on policy based resource sharing, so unit test and integration test is enough. 

## Queue

### Queue API (MVP)

Queue is the configuration of each tenant. Queue is a CDR (customer defined resource). The tools to query and update Queue is necessary (MVP). There'll be one Queue for each ResourceQuota:

1. User create both ResourceQuota and Queue manually (MVP)
2. kube-arbitrator create ResourceQuota if user only create Queue; the resource quota can not be patched.
3. kube-arbitrator create Queue if user only create ResourceQuota
4. if both ResourceQuota and Queue were not created, Queue admission controller will reject the Pod creation.

Currently, Queue is bound to Quota/Namespace; but there's discussion to make it cross namespace, e.g. [Multi-Project Quotas in OpenShift](https://docs.openshift.org/latest/admin_guide/multiproject_quota.html) . It dependent on the requirements from the workload controller, e.g. Spark; so we'd like to discuss this after MVP is done.

### Queue Controller

#### Objects cache (MVP)

Similar to kube-scheduler, kube-arbitrator need to cache related objects for algorithm. In kube-arbitrator need to cache Pod, Node and Queue for the algorithm. It need a cache to manage those objects.

#### DRF as default resource sharing policy (MVP)

DRF ([Dominant Resource Fairness](https://people.eecs.berkeley.edu/~alig/papers/drf.pdf)) is the default resource sharing policy; a policy interface is also required for other policy. The DRF not only account running Pods, but also account resource requirements.

In MVP, the resource requirements is configured in Queue, the resource requirement may be different from pod to pod. 

#### Resource re-shuffle (preemption) (MVP)

As DRF will re-calculate deserved resources of each tenant; if any tenant used more resource than its deserved resource, kube-arbitrator preempt some resources for it.

#### QueueQuota controller

#### Pluggable policy

#### Integration with default scheduler

### Queue Admission Controller

#### Reject pod creation if no Queue for its namespace (MVP)

Queue admission controller will reject pod creation if no Queue for its namespace; otherwise, it dependent on ResourceQuota admission controller.

#### QueueQuota admission controller cross namespace

## Integration with Spark (MVP)

Need @foxish's input

## QueueJob

### API

### Workload scheduling

### Integration with default scheduler

## ResourceRequest Estimation

## Hierarchical Namespace

## Others

### Oversubscription

### Data Aware Scheduling

### Integration with TensorFlow