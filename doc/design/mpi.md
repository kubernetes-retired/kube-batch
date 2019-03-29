## MPI Jobs in Kube-Batch

Kube-Batch has support for an MPI-style job. This combines PodGroups with additional Kubernetes primitives
to spin up and allow multiple pods to communicate between one another via [MPI](https://www.open-mpi.org/).

### How it works

An MPI object is comprised of a JobSpec and a PodGroupSpec. When an MPI object is created, it:

1. Creates a ServiceAccount, Role, and Rolebinding. This allows us granular control over what permissions
   the pods have, as well as supporting some cool NetworkPolicy features.
2. Creates two blank ConfigMap objects. One contains the MPI hostfile, the other contains data for
   /etc/hosts as well as the init scripts.
3. Creates a Job and a PodGroup keyed to the name of the MPI Object
4. Once all Pods have been created, Kube-Batch will update the ConfigMap objects with the IP addresses of
   all running MPI-Job Pods, as well as designate one Pod as the "Executor" -- the Pod that will actually
   run the desired command.
5. The Pod designated as the executor will download `kubectl` and use that as the transport channel to
   communicate with the other pods in the PodGroup. Once pod connection is established, it will begin
   running the MPI job.

### Caveats

Any image used in the JobSpec of the MPI Object must already have MPI support baked in. For our example,
we use `continuse/mpich:v3` as a base MPI image.

MPI jobs take some time to start once scheduled. This is due to, at runtime, downloading things like
`kubectl` and waiting for the state of all pods to be reconciled. There are many optimizations that
should be made over time.

### Example YAML

```yaml
apiVersion: scheduling.incubator.k8s.io/v1alpha1
kind: MPI
metadata:
  name: test-mpi
spec:
  job:
    apiVersion: batch/v1
    kind: Job
    spec:
      completions: 6
      parallelism: 6
      template:
        spec:
          containers:
          - image: continuse/mpich:v3
            imagePullPolicy: IfNotPresent
            name: mpi
            command: ["mpirun", "-n", "6", "hostname"]
            resources:
              requests:
                cpu: "200m"
  podGroup:
    spec:
      minMember: 6
```