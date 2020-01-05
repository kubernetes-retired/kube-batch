# Kube-batch scheduler framework

@ingvagabund; Jan 7, 2020

Kube-batch scheduler provides a framework which allows to simultaneously
schedule a group of tasks (e.g. all tasks must be first scheduled to available nodes
before each task starts to run). By integrating multi level queue scheduling
mechanism it allows to categorize and prioritize jobs (jobs in higher priority
queues will get processed first before jobs in lower priority queues).
Extensible through plugins the framework allows to define various conditions
for when a task can be preempted, what it means for a job to be ready, how to
sort jobs/tasks, how to prioritize multi-level queues, etc.
Logic responsible for the scheduling itself (including preemption) is implemented
through actions which build on top of the framework and use logic provided by the plugins.

## Framework

The framework provides various methods to be used within actions. E.g. `Reclaimable`,
`Preemptable`, `Overused`, `JobReady`, `JobPipelined`, `JobValid`, `JobOrderFn`,
`QueueOrderFn`, `TaskOrderFn`, `PredicateFn`, `NodePrioritizers`, `Pipeline`,
`Allocate` or `Evict`.

Many of the methods executes their logic in tiers. You can see one tier
as a list of plugins. The order of tiers is specified inside a configuration file.
Some methods stop iterating through the tier once they get first positive/negative
result (e.g. non-empty list of victims found, node predicate rejected a pod),
some methods want to iterate through all tiers (e.g. node predicates).

Both `Reclaimable` and `Preemptable` are computed in tiers. If no victims
(tasks to reclaim/preempt) in the current tier are found, next tier is processed.
In each tier, all reclaimable, resp. preemptable functions from the plugins
of the tier are executed, providing a list of potential victims. If intersection
of all the victim lists is non-empty, the intersection is returned. Otherwise,
next tier is processed until there's at least one victim or no more tiers left.

Similar holds for `Overused`, `JobReady`, `JobPipelined` and `JobValid`.
First tier that has at least one plugin giving positive, resp. non-empty result
for a given method wins. Remaining tiers are ignored. In case of `JobOrderFn`,
`QueueOrderFn` and `TaskOrderFn`, first tier with a plugin function that can tell
that two items are not equal wins (if two items are considered equal, next plugin
in the same tier is called, resp. plugin in the next tier is if no more plugins
in the same tier are available). In case both items are still considered equal
and there's no more tiers left, creation timestamp is taken into account (or
item's UID as the last resort).

In case of `PredicateFn`, predicate functions of all plugins of all tiers are executed.
In case there's one function that returns an error, overall result is the error.
Otherwise nil. In case of `NodePrioritizers`, all priority configs from all
plugins of all tiers are appended and returned.

The framework also carries a cache data type that contains kubernetes clients,
informers and communicates with the API server. E.g. anytime the framework needs
to preempt or evict a pod, the request boils down to calling `CoreV1().Pods().Delete()`,
resp. `CoreV1().Pods().Bind()`. The framework exposes corresponding `Binder.Bind`
and `Evictor.Evict` functions of the cache that can be called.

`Allocate` allocates volumes, set tasks status to `Allocated` and update
the framework's node cache. Once all is successfully done and corresponding job
is considered ready, all allocated volumes and tasks are bound (commited into the cluster).
Tasks's status is changed to `Binding` and the node cache updated again. `Pipeline`,
compared to `Allocate`, only updates tasks's status and node cache without committing
the changes into the cluster. Both `Allocated` and `Pipeline` triggers `AllocateFunc`
event handlers in case of success. `Evict` first deletes tasks's pod, updates
task's status, node info and triggers all `DeallocateFunc` event handlers in case of success.

There's also a concept of a `Statement` with `Commit` and `Discard` methods that
allows to postpone execution of changes into the cluster. The data type allows
to make a list of evict and/or pipeline commands while committing changes only
in the cache. In case of invoking `Commit`, all changes gets committed into the cluster.
In case of invoking `Discard`, all changes in the cache gets undone without committing
anything into the cluster.

## Plugins

### Conformance plugin

The plugin registers implementation of a function deciding whether a task is evictable.
Framework's methods utilizing the plugin:
- `Preemptable`
- `Reclaimable`

### DRF plugin

The plugin computes a resource score for each job.
Job score corresponds to a maximum of all resource ratios. Resource ration is
consumption of a given resource (e.g. cpu, memory) through all tasks of a job
divided by allocatable amount of the resource through all nodes.

The plugin registers implementation of:
- a function deciding whether a preemptor replacing a preemptee will result in better resource distribution among jobs (for `Preemptable`)
- a function ordering jobs based on their score from the smallest to the largest.

Also registers event handlers for recomputing the score in case a task is allocated/deallocated.

Framework's methods utilizing the plugin:
- `Preemptable`
- `JobOrderFn`

### Gang plugin

Registers a function validating a job for gang scheduling.
The function rejects a job when a number of valid tasks is less than a number of minimal
tasks required by the job.

The plugin registers implementation of:
- a function that forbids preemption in case the number of minimal tasks would be violated
- a job ordering function ordering jobs from not ready to ready (a job is ready when the minimal number of tasks is ready).

Framework's methods utilizing the plugin:
- `JobValidFn`
- `Reclaimable`
- `Preemptable`
- `JobOrderFn`
- `JobReadyFn`
- `JobPipelinedFn`

### Nodeorder plugin

Registers node prioritizer. Code of individual node prioritizers is imported from the scheduler framework. Available node priorities:
- `NodeAffinityPriority`
- `InterPodAffinityPriority`
- `LeastRequestedPriority`
- `MostRequestedPriority`
- `BalancedResourceAllocation`

Framework's methods utilizing the plugin:
- `NodePrioritizers`

### Predicates plugin

Registers node predicates. Code of individual node predicates is imported from the scheduler framework. Namely:
- `CheckNodeConditionPredicate`
- `CheckNodeUnschedulablePredicate`
- `PodMatchNodeSelector`
- `PodFitsHostPorts`
- `PodToleratesNodeTaints`
- `CheckNodeMemoryPressurePredicate`
- `CheckNodeDiskPressurePredicate`
- `CheckNodePIDPressurePredicate`
- `NewPodAffinityPredicate`

Framework's methods utilizing the plugin:
- `PredicateFn`

### Priority plugin

The plugin registers implementation of:
- a task ordering function based on task priority: higher priority tasks before lower priority tasks
- a job ordering function based on job priority: higher priority jobs before lower priority jobs
- preemption decision logic: task of a job with a lower priority can not preempt task of a job with higher priority

Framework's methods utilizing the plugin:
- `TaskOrderFn`
- `JobOrderFn`
- `Preemptable`

### Proportion plugin

The plugin computes deserved shares of all queues based on their weight
and number of jobs that would be successfully scheduled.
First, all queue get assigned a share solely based on their ratio of their weight
and the total weight. If resources share assigned to a queue is sufficient to cover
requested resources of all jobs in the queue, the queue is considered served
and its share is pinned. Any abundant resources (e.g. cpu, memory) not used by jobs
in the pinned queue is returned to the remaining resource pool. In the second
and the next run the remaining resources are again distributed among queues
that still do not have enough resources to satisfy resource requests of all
their jobs (in the same weight ration each queue has). Eventually, either all
remaining resources are depleted or all queues have enough resources to cover
resource requests of all their jobs. Queues shares are then updated accordingly.

The resource shares then serve as a base for implementing:
- queue order function: queues with smaller shares are less than queues with larger shares
- reclaimable function: a resource from a queue can be reclaimed if a queue still has at least the same share of allocated resources even after removing a task
- overuses function: true when a queue has more allocated resources than it deserves (taking more than its share)

Also registers event handlers for recomputing the shares in case a task is allocated/deallocated.

Framework's methods utilizing the plugin:
- `QueueOrderFn`
- `Reclaimable`
- `Overused`

## Actions

### Allocate action

Action responsible for batch scheduling of jobs (collection of tasks) onto nodes.
Every time the action is executed, it reads all the available jobs and constructs
multi-level queues based on each job category. Priority for sorting the multi-level queues
depends on which plugins are enabled. Jobs from the higher-priority queues
are processed first. Each processing reads all job's tasks that are still pending
and tries to find a set of nodes (with sufficient resources) that can run
the tasks (using predicates and priorities in the same manner as the kubernetes scheduler).
Also, if a job is considered ready (depends on which plugins are enabled),
tasks scheduling is interrupted and the job is pushed back into the job queue
and re-processed so jobs with higher priority can be processed sooner (some jobs
can have their priority decreased with increased number of tasks running).
If the higher-priority queue has no jobs left (or no tasks can be scheduled,
e.g. due to insufficient resources), next highest-priority queue is processed.

Info: the action ignores any task with empty Resreq

### Backfill action

Action responsible for scheduling tasks with empty InitResreq.
It goes through a list of nodes and picks first node complying with all configured
predicates. Node priority is ignored and marked as TODO.

There seems to be an overlap with the allocate action. As long as both action
are not run in parallel, there's no conflict of interest.

### Preempt action

Action responsible for preempting tasks from nodes based on their priority
compared to already running tasks.

The action first constructs a list of all pending tasks (called preemptor tasks).
Then, for each preemptor task the action finds the most suitable node (based
on predicates and priorities) and decides if there are tasks on the node that
can be preempted by the preemptor. If not, the next most suitable node is picked
and the processes is repeated.

The preemption itself first marks a task to be in Releasing state (including
updating node info about the state change). Once a sufficient amount of resources
is freed for the premptor task to be run, premptor task is scheduled and the next
preemptor tasks is processed. If a job has no more preemptors, next job is processed.
The preemptee tasks are sorted before preemption so the least priority tasks are terminated first.

Each queue from the multi-level queues is processed in two phases. In the first,
only tasks from the same queue (jobs with the same priority) are processed (also
ignoring tasks from the same job as the preemptor tasks is).
In the second, only tasks from within the same job are processed. Just in case
the first phase still leave some pending tasks which can potentially preempt tasks
within the same job (in the worst case).

### Reclaim action

Reclaim action is responsible for terminating tasks with the same intent as Preempt action.
To free resources for high-priority tasks/jobs. The difference is in the order
of processing jobs and their scope. In this case, jobs from higher priority queues
are processed first. Preempting tasks of jobs from queues with lower priority
than queue of a preemptor task. If a job from a queue can't be scheduled
(after the reclamation), next job is processed. If there are no more jobs left
in a current high priority queue to be processed, jobs from a queue with the next
highest priority are processed.
