## v0.5  

**Closed issues:**

- master and worker start in order in a job [\#872](https://github.com/kubernetes-sigs/kube-batch/issues/872)
-  use kubernetes default schedule policies [\#864](https://github.com/kubernetes-sigs/kube-batch/issues/864)
- v0.4 runtime panic  [\#861](https://github.com/kubernetes-sigs/kube-batch/issues/861)
- Helm install fails [\#857](https://github.com/kubernetes-sigs/kube-batch/issues/857)
- "NewResource/Convert2K8sResource" behavior mismatch [\#851](https://github.com/kubernetes-sigs/kube-batch/issues/851)
- Phase of podgroup status looks incorrect [\#846](https://github.com/kubernetes-sigs/kube-batch/issues/846)
- Resolve all golint issues ignored in golint failure file [\#837](https://github.com/kubernetes-sigs/kube-batch/issues/837)
- Fix e2e test: statement [\#836](https://github.com/kubernetes-sigs/kube-batch/issues/836)
- Fix e2e test: gang scheduling. [\#835](https://github.com/kubernetes-sigs/kube-batch/issues/835)
- Update API GroupName [\#815](https://github.com/kubernetes-sigs/kube-batch/issues/815)
- Migrate nodeorder and predicates plugins [\#814](https://github.com/kubernetes-sigs/kube-batch/issues/814)
- some files are repeated in multiple folders [\#804](https://github.com/kubernetes-sigs/kube-batch/issues/804)
- the job cannot be scheduled [\#803](https://github.com/kubernetes-sigs/kube-batch/issues/803)
- Add Configuration for predicate plugin to enable/disable predicates algorithm [\#802](https://github.com/kubernetes-sigs/kube-batch/issues/802)
- Propose not to delete all pods after the job finished [\#797](https://github.com/kubernetes-sigs/kube-batch/issues/797)
- verify-gencode.sh can not work [\#788](https://github.com/kubernetes-sigs/kube-batch/issues/788)
- Consider support multi-containers pod error code handling [\#776](https://github.com/kubernetes-sigs/kube-batch/issues/776)
- vkctrl enhancement [\#774](https://github.com/kubernetes-sigs/kube-batch/issues/774)
- Removed reclaim & preempt actions in vk-scheduler \(kube-batch\) [\#770](https://github.com/kubernetes-sigs/kube-batch/issues/770)
- Support Task/Job retry [\#769](https://github.com/kubernetes-sigs/kube-batch/issues/769)
- Job GC [\#768](https://github.com/kubernetes-sigs/kube-batch/issues/768)
- Support PriorityClassName in Job Resource [\#766](https://github.com/kubernetes-sigs/kube-batch/issues/766)
- Support ScheduledJob/CronJob  [\#765](https://github.com/kubernetes-sigs/kube-batch/issues/765)
- Queue controller & cli [\#763](https://github.com/kubernetes-sigs/kube-batch/issues/763)
- Update related doc in kubeflow [\#761](https://github.com/kubernetes-sigs/kube-batch/issues/761)
- Combine kube-batch & volcano e2e tests [\#760](https://github.com/kubernetes-sigs/kube-batch/issues/760)
- Deserved attr is not correctly calculated in proportion plugin [\#729](https://github.com/kubernetes-sigs/kube-batch/issues/729)
- Keep backward compatibility for priority class [\#724](https://github.com/kubernetes-sigs/kube-batch/issues/724)
- Invitation to Slack  [\#710](https://github.com/kubernetes-sigs/kube-batch/issues/710)
- Change return value of NodeOrderFn from int to float [\#708](https://github.com/kubernetes-sigs/kube-batch/issues/708)
- Add type Argument with some common parse function [\#704](https://github.com/kubernetes-sigs/kube-batch/issues/704)
- Replace NodeOrder with BestNode [\#699](https://github.com/kubernetes-sigs/kube-batch/issues/699)
- Support set default value to the configuration [\#695](https://github.com/kubernetes-sigs/kube-batch/issues/695)
- Add resource predicates for tasks [\#694](https://github.com/kubernetes-sigs/kube-batch/issues/694)
- Release 0.4.2 [\#672](https://github.com/kubernetes-sigs/kube-batch/issues/672)
- \[Question\]  Some confusion about Queue, podGroup [\#670](https://github.com/kubernetes-sigs/kube-batch/issues/670)
- \[Question\] Support ignore resource in some namespaces? [\#661](https://github.com/kubernetes-sigs/kube-batch/issues/661)
- Pass conformance test [\#589](https://github.com/kubernetes-sigs/kube-batch/issues/589)
- big PodGroup blocks scheduling issue [\#514](https://github.com/kubernetes-sigs/kube-batch/issues/514)
- Changing the architecture diagram [\#504](https://github.com/kubernetes-sigs/kube-batch/issues/504)
- Support MPI job by kube-batch [\#351](https://github.com/kubernetes-sigs/kube-batch/issues/351)
- Currently no IBM Power\(ppc64le\) based image is available in kubearbitrator public repository [\#241](https://github.com/kubernetes-sigs/kube-batch/issues/241)
- Add golint check for source code [\#64](https://github.com/kubernetes-sigs/kube-batch/issues/64)

**Merged pull requests:**

- Remove the useless namespace [\#877](https://github.com/kubernetes-sigs/kube-batch/pull/877) ([y-taka-23](https://github.com/y-taka-23))
- Add Unisound to who-is-using doc [\#873](https://github.com/kubernetes-sigs/kube-batch/pull/873) ([xieydd](https://github.com/xieydd))
- Add v1alpha2 Version for queue [\#871](https://github.com/kubernetes-sigs/kube-batch/pull/871) ([thandayuthapani](https://github.com/thandayuthapani))
- Support Both versions of PodGroup [\#870](https://github.com/kubernetes-sigs/kube-batch/pull/870) ([thandayuthapani](https://github.com/thandayuthapani))
- update  tutorial.md [\#865](https://github.com/kubernetes-sigs/kube-batch/pull/865) ([davidstack](https://github.com/davidstack))
- \[Cherry-Pick \#26\] Ignore nodes if out of syc. [\#863](https://github.com/kubernetes-sigs/kube-batch/pull/863) ([asifdxtreme](https://github.com/asifdxtreme))
- Migrate from dep to go modules to manage dependencies. [\#862](https://github.com/kubernetes-sigs/kube-batch/pull/862) ([haosdent](https://github.com/haosdent))
- Fix wrong calculation for nodeinfo used [\#860](https://github.com/kubernetes-sigs/kube-batch/pull/860) ([wackxu](https://github.com/wackxu))
- update node info no matter what node info changed [\#859](https://github.com/kubernetes-sigs/kube-batch/pull/859) ([wackxu](https://github.com/wackxu))
- Add Wait time for KIND cluster to become ready [\#856](https://github.com/kubernetes-sigs/kube-batch/pull/856) ([thandayuthapani](https://github.com/thandayuthapani))
- support priorities map-reduce [\#855](https://github.com/kubernetes-sigs/kube-batch/pull/855) ([wackxu](https://github.com/wackxu))
-  Added Configuration to enable/disable predicates [\#854](https://github.com/kubernetes-sigs/kube-batch/pull/854) ([thandayuthapani](https://github.com/thandayuthapani))
- Updating running condition for pod group phase  [\#850](https://github.com/kubernetes-sigs/kube-batch/pull/850) ([nikita15p](https://github.com/nikita15p))
- Add Comment in allocate [\#849](https://github.com/kubernetes-sigs/kube-batch/pull/849) ([thandayuthapani](https://github.com/thandayuthapani))
- Use KIND instead of DIND & Include Skipped test cases [\#848](https://github.com/kubernetes-sigs/kube-batch/pull/848) ([thandayuthapani](https://github.com/thandayuthapani))
- Revert Vkctl queue subcommand [\#844](https://github.com/kubernetes-sigs/kube-batch/pull/844) ([Rajadeepan](https://github.com/Rajadeepan))
- Changing the event source to the right scheduler name [\#843](https://github.com/kubernetes-sigs/kube-batch/pull/843) ([shivramsrivastava](https://github.com/shivramsrivastava))
- Added Queue Capability. [\#841](https://github.com/kubernetes-sigs/kube-batch/pull/841) ([k82cn](https://github.com/k82cn))
- Fixed the Pod failure event reason string for the k8s conformance tes… [\#839](https://github.com/kubernetes-sigs/kube-batch/pull/839) ([shivramsrivastava](https://github.com/shivramsrivastava))
- Added user list in README. [\#834](https://github.com/kubernetes-sigs/kube-batch/pull/834) ([k82cn](https://github.com/k82cn))
- Add Vivo to who-is-using doc [\#833](https://github.com/kubernetes-sigs/kube-batch/pull/833) ([zionwu](https://github.com/zionwu))
- Add New API Group for PodGroup [\#832](https://github.com/kubernetes-sigs/kube-batch/pull/832) ([thandayuthapani](https://github.com/thandayuthapani))
- Replace gofmt with goimports [\#831](https://github.com/kubernetes-sigs/kube-batch/pull/831) ([hex108](https://github.com/hex108))
- Adding subcommand queue create and list to vkctl [\#820](https://github.com/kubernetes-sigs/kube-batch/pull/820) ([Rajadeepan](https://github.com/Rajadeepan))
- Add script for e2e prow job [\#819](https://github.com/kubernetes-sigs/kube-batch/pull/819) ([asifdxtreme](https://github.com/asifdxtreme))
- Fix kubemark issue & Add document for kubemark test [\#817](https://github.com/kubernetes-sigs/kube-batch/pull/817) ([TommyLike](https://github.com/TommyLike))
- Migrate duplicated code in nodeorder and predicate plugin [\#816](https://github.com/kubernetes-sigs/kube-batch/pull/816) ([thandayuthapani](https://github.com/thandayuthapani))
- Added Queue status. [\#812](https://github.com/kubernetes-sigs/kube-batch/pull/812) ([k82cn](https://github.com/k82cn))
- Automated cherry pick of \#809: Prevent memory leak [\#811](https://github.com/kubernetes-sigs/kube-batch/pull/811) ([k82cn](https://github.com/k82cn))
- Add kubemark support [\#810](https://github.com/kubernetes-sigs/kube-batch/pull/810) ([TommyLike](https://github.com/TommyLike))
- Prevent memory leak [\#809](https://github.com/kubernetes-sigs/kube-batch/pull/809) ([hzxuzhonghu](https://github.com/hzxuzhonghu))
- Update tutorial.md [\#808](https://github.com/kubernetes-sigs/kube-batch/pull/808) ([hzxuzhonghu](https://github.com/hzxuzhonghu))
- Automated cherry pick of \#725: add options to enable the feature of PriorityClass [\#807](https://github.com/kubernetes-sigs/kube-batch/pull/807) ([k82cn](https://github.com/k82cn))
- Revert volcano changes [\#806](https://github.com/kubernetes-sigs/kube-batch/pull/806) ([TommyLike](https://github.com/TommyLike))
- Add UT Cases for Reclaim Action [\#801](https://github.com/kubernetes-sigs/kube-batch/pull/801) ([thandayuthapani](https://github.com/thandayuthapani))
- Add CheckNodeDiskPressure,CheckNodePIDPressure,CheckNodeMemoryPressur… [\#800](https://github.com/kubernetes-sigs/kube-batch/pull/800) ([wackxu](https://github.com/wackxu))
- Add Volcano Installation Tutorial Link [\#796](https://github.com/kubernetes-sigs/kube-batch/pull/796) ([asifdxtreme](https://github.com/asifdxtreme))
- support extended resource for kube-batch [\#795](https://github.com/kubernetes-sigs/kube-batch/pull/795) ([wackxu](https://github.com/wackxu))
- Add document for MPI example [\#793](https://github.com/kubernetes-sigs/kube-batch/pull/793) ([TommyLike](https://github.com/TommyLike))
- Fix verify [\#792](https://github.com/kubernetes-sigs/kube-batch/pull/792) ([hzxuzhonghu](https://github.com/hzxuzhonghu))
- restore kube-batch [\#791](https://github.com/kubernetes-sigs/kube-batch/pull/791) ([hzxuzhonghu](https://github.com/hzxuzhonghu))
- Fix Broken image link [\#787](https://github.com/kubernetes-sigs/kube-batch/pull/787) ([Rajadeepan](https://github.com/Rajadeepan))
- Fix Typo [\#785](https://github.com/kubernetes-sigs/kube-batch/pull/785) ([Rajadeepan](https://github.com/Rajadeepan))
- Add generate code tools. [\#784](https://github.com/kubernetes-sigs/kube-batch/pull/784) ([TommyLike](https://github.com/TommyLike))
- Keep kube-batch doc compatible [\#783](https://github.com/kubernetes-sigs/kube-batch/pull/783) ([TommyLike](https://github.com/TommyLike))
- Fix typo [\#782](https://github.com/kubernetes-sigs/kube-batch/pull/782) ([jimbobby5](https://github.com/jimbobby5))
- Help message for vkctl job [\#779](https://github.com/kubernetes-sigs/kube-batch/pull/779) ([Rajadeepan](https://github.com/Rajadeepan))
- fix delete po err msg [\#777](https://github.com/kubernetes-sigs/kube-batch/pull/777) ([wangyuqing4](https://github.com/wangyuqing4))
- Implement error code handling in lifecyclePolicy [\#775](https://github.com/kubernetes-sigs/kube-batch/pull/775) ([hzxuzhonghu](https://github.com/hzxuzhonghu))
- Abstract a common pod delete func [\#773](https://github.com/kubernetes-sigs/kube-batch/pull/773) ([hzxuzhonghu](https://github.com/hzxuzhonghu))
- Remove reclaim and preempt in deployment [\#772](https://github.com/kubernetes-sigs/kube-batch/pull/772) ([thandayuthapani](https://github.com/thandayuthapani))
- Add hzxuzhonghu to approver & reviewers [\#771](https://github.com/kubernetes-sigs/kube-batch/pull/771) ([hzxuzhonghu](https://github.com/hzxuzhonghu))
- Added volcano logo. [\#767](https://github.com/kubernetes-sigs/kube-batch/pull/767) ([k82cn](https://github.com/k82cn))
- Combine volcano & kube batch e2e tests [\#764](https://github.com/kubernetes-sigs/kube-batch/pull/764) ([TommyLike](https://github.com/TommyLike))
- Add document for e2e tests [\#759](https://github.com/kubernetes-sigs/kube-batch/pull/759) ([TommyLike](https://github.com/TommyLike))
- Updated README accordingly. [\#758](https://github.com/kubernetes-sigs/kube-batch/pull/758) ([k82cn](https://github.com/k82cn))
- Renamed kube-batch to Volcano [\#757](https://github.com/kubernetes-sigs/kube-batch/pull/757) ([k82cn](https://github.com/k82cn))
- \[Add Job Resource\] Fix some issues when enabling e2e tests [\#756](https://github.com/kubernetes-sigs/kube-batch/pull/756) ([TommyLike](https://github.com/TommyLike))
- Removed inactive owners. [\#754](https://github.com/kubernetes-sigs/kube-batch/pull/754) ([k82cn](https://github.com/k82cn))
- Enabled e2e-kind. [\#753](https://github.com/kubernetes-sigs/kube-batch/pull/753) ([k82cn](https://github.com/k82cn))
- \[Add Job Resource\] Fix panic error when running job e2e tests [\#752](https://github.com/kubernetes-sigs/kube-batch/pull/752) ([TommyLike](https://github.com/TommyLike))
- Add UT cases for Preempt Action [\#751](https://github.com/kubernetes-sigs/kube-batch/pull/751) ([thandayuthapani](https://github.com/thandayuthapani))
- \[Add Job Resource\] Support build&e2e tests workflow [\#750](https://github.com/kubernetes-sigs/kube-batch/pull/750) ([TommyLike](https://github.com/TommyLike))
- Delete binary controller [\#747](https://github.com/kubernetes-sigs/kube-batch/pull/747) ([hex108](https://github.com/hex108))
- \[Support Job Resource\] Add related document [\#746](https://github.com/kubernetes-sigs/kube-batch/pull/746) ([TommyLike](https://github.com/TommyLike))
- Fix Lint errors for pkg/scheduler/api package [\#745](https://github.com/kubernetes-sigs/kube-batch/pull/745) ([thandayuthapani](https://github.com/thandayuthapani))
- FixGolint in scheduler package [\#744](https://github.com/kubernetes-sigs/kube-batch/pull/744) ([Rajadeepan](https://github.com/Rajadeepan))
- \[Support Job Resource\] Add related helm chart [\#743](https://github.com/kubernetes-sigs/kube-batch/pull/743) ([TommyLike](https://github.com/TommyLike))
- Fix Golint Errors in pkg/controllers/job package [\#742](https://github.com/kubernetes-sigs/kube-batch/pull/742) ([thandayuthapani](https://github.com/thandayuthapani))
- \[Support Job Resource\] Add volcano e2e [\#741](https://github.com/kubernetes-sigs/kube-batch/pull/741) ([hzxuzhonghu](https://github.com/hzxuzhonghu))
- Fix GoLint Errors in pkg/controllers/apis and pkg/controllers/cache packages [\#740](https://github.com/kubernetes-sigs/kube-batch/pull/740) ([thandayuthapani](https://github.com/thandayuthapani))
- Fix Golint Errors in cmd package [\#739](https://github.com/kubernetes-sigs/kube-batch/pull/739) ([Rajadeepan](https://github.com/Rajadeepan))
- Fix Golint Errors for pkg/admission and pkg/cli/job package [\#738](https://github.com/kubernetes-sigs/kube-batch/pull/738) ([thandayuthapani](https://github.com/thandayuthapani))
- fix verify [\#737](https://github.com/kubernetes-sigs/kube-batch/pull/737) ([hzxuzhonghu](https://github.com/hzxuzhonghu))
- Built binaries. [\#734](https://github.com/kubernetes-sigs/kube-batch/pull/734) ([k82cn](https://github.com/k82cn))
- \[Support Job Resource\] Add vkctl [\#733](https://github.com/kubernetes-sigs/kube-batch/pull/733) ([hzxuzhonghu](https://github.com/hzxuzhonghu))
- \[Support Job Resource\] Add job controller [\#732](https://github.com/kubernetes-sigs/kube-batch/pull/732) ([hzxuzhonghu](https://github.com/hzxuzhonghu))
- \[Support Job Resource\] Add job admission  [\#731](https://github.com/kubernetes-sigs/kube-batch/pull/731) ([TommyLike](https://github.com/TommyLike))
- Fix wrong calculation for queue deserved in proportion plugin [\#730](https://github.com/kubernetes-sigs/kube-batch/pull/730) ([zionwu](https://github.com/zionwu))
- \[Support Job Resource\] Add basic job definition [\#728](https://github.com/kubernetes-sigs/kube-batch/pull/728) ([TommyLike](https://github.com/TommyLike))
- \[DOCS\] Docs for Reclaim Action [\#727](https://github.com/kubernetes-sigs/kube-batch/pull/727) ([thandayuthapani](https://github.com/thandayuthapani))
- add options to enable the feature of PriorityClass [\#725](https://github.com/kubernetes-sigs/kube-batch/pull/725) ([jiaxuanzhou](https://github.com/jiaxuanzhou))
- \[DOCS\] Docs for Preempt action [\#723](https://github.com/kubernetes-sigs/kube-batch/pull/723) ([thandayuthapani](https://github.com/thandayuthapani))
- Adding spell checker and boilerplate verify script [\#722](https://github.com/kubernetes-sigs/kube-batch/pull/722) ([shivramsrivastava](https://github.com/shivramsrivastava))
- Select only one best node for allocate [\#721](https://github.com/kubernetes-sigs/kube-batch/pull/721) ([hex108](https://github.com/hex108))
- NodeOrderFn return float value [\#719](https://github.com/kubernetes-sigs/kube-batch/pull/719) ([lmzqwer2](https://github.com/lmzqwer2))
- Added resource predicates. [\#717](https://github.com/kubernetes-sigs/kube-batch/pull/717) ([k82cn](https://github.com/k82cn))
- Improve preempt performance by predicate and priority concurrently [\#716](https://github.com/kubernetes-sigs/kube-batch/pull/716) ([hex108](https://github.com/hex108))
- \[DOCS\] Add workflow of Reclaim action [\#715](https://github.com/kubernetes-sigs/kube-batch/pull/715) ([thandayuthapani](https://github.com/thandayuthapani))
- \[DOCS\] Add the Execution Workflow for Allocation of Jobs to Nodes [\#714](https://github.com/kubernetes-sigs/kube-batch/pull/714) ([asifdxtreme](https://github.com/asifdxtreme))
- removing unnecessary entries and syntax changes [\#713](https://github.com/kubernetes-sigs/kube-batch/pull/713) ([animeshsingh](https://github.com/animeshsingh))
- fixing grammatical errors [\#712](https://github.com/kubernetes-sigs/kube-batch/pull/712) ([animeshsingh](https://github.com/animeshsingh))
- distinguishing the key words in batch scheduler [\#711](https://github.com/kubernetes-sigs/kube-batch/pull/711) ([animeshsingh](https://github.com/animeshsingh))
- Fixed word issue in plugin-conf document. [\#709](https://github.com/kubernetes-sigs/kube-batch/pull/709) ([yylin1](https://github.com/yylin1))
- Docs for task-order within Job [\#706](https://github.com/kubernetes-sigs/kube-batch/pull/706) ([thandayuthapani](https://github.com/thandayuthapani))
- Add type Argument with some common parse function [\#705](https://github.com/kubernetes-sigs/kube-batch/pull/705) ([lmzqwer2](https://github.com/lmzqwer2))
- Docs for DRF Plugin [\#703](https://github.com/kubernetes-sigs/kube-batch/pull/703) ([Rajadeepan](https://github.com/Rajadeepan))
- Fixed table issue in tutorial. [\#702](https://github.com/kubernetes-sigs/kube-batch/pull/702) ([k82cn](https://github.com/k82cn))
- Apply for approver [\#701](https://github.com/kubernetes-sigs/kube-batch/pull/701) ([hex108](https://github.com/hex108))
- Fix race condition for concurrent predicate and priority in allocate [\#700](https://github.com/kubernetes-sigs/kube-batch/pull/700) ([hex108](https://github.com/hex108))
- Add support for setting default value to the configuration [\#698](https://github.com/kubernetes-sigs/kube-batch/pull/698) ([hex108](https://github.com/hex108))
- Improve allocate performance by making predicate and priority concurr… [\#696](https://github.com/kubernetes-sigs/kube-batch/pull/696) ([hex108](https://github.com/hex108))
- Update helm chart version to 0.4.2 [\#693](https://github.com/kubernetes-sigs/kube-batch/pull/693) ([Jeffwan](https://github.com/Jeffwan))
- Improve kube-batch documentation [\#668](https://github.com/kubernetes-sigs/kube-batch/pull/668) ([Jeffwan](https://github.com/Jeffwan))
- Do not count in pipelined task when calculating ready tasks number [\#650](https://github.com/kubernetes-sigs/kube-batch/pull/650) ([hex108](https://github.com/hex108))






## v0.4.2

### Notes:
 * Added Balanced Resource Priority
 * Support Plugin Arguments
 * Support Weight for Node order plugin

### Issues:
		
 * [#636](https://github.com/kubernetes-sigs/kube-batch/pull/636)	add balanced resource priority	([@DeliangFan](https://github.com/DeliangFan))
 * [#638](https://github.com/kubernetes-sigs/kube-batch/pull/638)	Take init containers into account when getting pod resource request	([@hex108](https://github.com/hex108))
 * [#639](https://github.com/kubernetes-sigs/kube-batch/pull/639)	Support Plugin Arguments	([@thandayuthapani](https://github.com/thandayuthapani))
 * [#640](https://github.com/kubernetes-sigs/kube-batch/pull/640)	Support Weight for NodeOrder Plugin	([@thandayuthapani](https://github.com/thandayuthapani))
 * [#642](https://github.com/kubernetes-sigs/kube-batch/pull/642)	Add event when task is scheduled	([@Rajadeepan](https://github.com/Rajadeepan))
 * [#643](https://github.com/kubernetes-sigs/kube-batch/pull/643)	Return err in Allocate if any error occurs	([@hex108](https://github.com/hex108))
 * [#644](https://github.com/kubernetes-sigs/kube-batch/pull/644)	Fix Typos	([@TommyLike](https://github.com/TommyLike))
 * [#645](https://github.com/kubernetes-sigs/kube-batch/pull/645)	Order task by CreationTimestamp first, then by UID	([@hex108](https://github.com/hex108))
 * [#647](https://github.com/kubernetes-sigs/kube-batch/pull/647)	In allocate, skip adding Job if its queue is not found	([@hex108](https://github.com/hex108))
 * [#649](https://github.com/kubernetes-sigs/kube-batch/pull/649)	Preempt lowest priority task first	([@hex108](https://github.com/hex108))
 * [#651](https://github.com/kubernetes-sigs/kube-batch/pull/651)	Return err in functions of session.go if any error occurs	([@hex108](https://github.com/hex108))
 * [#652](https://github.com/kubernetes-sigs/kube-batch/pull/652)	Change run option SchedulePeriod's type to make it clear	([@hex108](https://github.com/hex108))
 * [#655](https://github.com/kubernetes-sigs/kube-batch/pull/655)	Do graceful eviction using default policy	([@hex108](https://github.com/hex108))
 * [#658](https://github.com/kubernetes-sigs/kube-batch/pull/658)	Address helm install error in tutorial.md	([@hex108](https://github.com/hex108))
 * [#660](https://github.com/kubernetes-sigs/kube-batch/pull/660)	Fix sub exception in reclaim and preempt	([@TommyLike](https://github.com/TommyLike))
 * [#666](https://github.com/kubernetes-sigs/kube-batch/pull/666)	Fix wrong caculation for deserved in proportion plugin	([@zionwu](https://github.com/zionwu))
 * [#671](https://github.com/kubernetes-sigs/kube-batch/pull/671)	Change base image to alphine to reduce image size	([@hex108](https://github.com/hex108))
 * [#673](https://github.com/kubernetes-sigs/kube-batch/pull/673)	Do not create PodGroup and Job for task whose scheduler is not kube-batch	([@hex108](https://github.com/hex108))


## v0.4.1

### Notes:

  * Added NodeOrder Plugin
  * Added Conformance Plugin
  * Removed namespaceAsQueue feature
  * Supported Pod without PodGroup
  * Added performance metrics for scheduling
  
### Issues:

  * [#632](https://github.com/kubernetes-sigs/kube-batch/pull/632) Set schedConf to defaultSchedulerConf if failing to readSchedulerConf ([@hex108](https://github.com/hex108))
  * [#631](https://github.com/kubernetes-sigs/kube-batch/pull/631) Add helm hook crd-install to fix helm install error ([@hex108](https://github.com/hex108))
  * [#584](https://github.com/kubernetes-sigs/kube-batch/pull/584) Replaced FIFO by workqueue. ([@k82cn](https://github.com/k82cn))
  * [#585](https://github.com/kubernetes-sigs/kube-batch/pull/585) Removed invalid error log. ([@k82cn](https://github.com/k82cn))
  * [#594](https://github.com/kubernetes-sigs/kube-batch/pull/594) Fixed Job.Clone issue. ([@k82cn](https://github.com/k82cn))
  * [#600](https://github.com/kubernetes-sigs/kube-batch/pull/600) Fixed duplicated queue in preempt action. ([@k82cn](https://github.com/k82cn))
  * [#596](https://github.com/kubernetes-sigs/kube-batch/pull/596) Set default PodGroup for Pods. ([@k82cn](https://github.com/k82cn))
  * [#587](https://github.com/kubernetes-sigs/kube-batch/pull/587) Adding node priority ([@thandayuthapani](https://github.com/thandayuthapani))
  * [#607](https://github.com/kubernetes-sigs/kube-batch/pull/607) Fixed flaky test. ([@k82cn](https://github.com/k82cn))
  * [#610](https://github.com/kubernetes-sigs/kube-batch/pull/610) Updated log level. ([@k82cn](https://github.com/k82cn))
  * [#609](https://github.com/kubernetes-sigs/kube-batch/pull/609) Added conformance plugin. ([@k82cn](https://github.com/k82cn))
  * [#604](https://github.com/kubernetes-sigs/kube-batch/pull/604) Moved default implementation from pkg. ([@k82cn](https://github.com/k82cn))
  * [#613](https://github.com/kubernetes-sigs/kube-batch/pull/613) Removed namespaceAsQueue. ([@k82cn](https://github.com/k82cn))
  * [#615](https://github.com/kubernetes-sigs/kube-batch/pull/615) Handled minAvailable task everytime. ([@k82cn](https://github.com/k82cn))
  * [#603](https://github.com/kubernetes-sigs/kube-batch/pull/603) Added PriorityClass to PodGroup. ([@k82cn](https://github.com/k82cn))
  * [#618](https://github.com/kubernetes-sigs/kube-batch/pull/618) Added JobPriority e2e. ([@k82cn](https://github.com/k82cn))
  * [#622](https://github.com/kubernetes-sigs/kube-batch/pull/622) Update doc & deployment. ([@k82cn](https://github.com/k82cn))
  * [#626](https://github.com/kubernetes-sigs/kube-batch/pull/626) Fix helm deployment ([@thandayuthapani](https://github.com/thandayuthapani))
  * [#625](https://github.com/kubernetes-sigs/kube-batch/pull/625) Rename PrioritizeFn to NodeOrderFn ([@thandayuthapani](https://github.com/thandayuthapani))
  * [#592](https://github.com/kubernetes-sigs/kube-batch/pull/592) Add performance metrics for scheduling ([@Jeffwan](https://github.com/Jeffwan))

## v0.4

### Notes:

  * Gang-scheduling/Coscheduling by PDB is depreciated.
  * Scheduler configuration format and related start up parameter is updated, refer to [example](https://github.com/kubernetes-sigs/kube-batch/blob/release-0.4/example/kube-batch-conf.yaml) for detail.

### Issues:

  * [#534](https://github.com/kubernetes-sigs/kube-batch/pull/534) Removed old design doc to avoid confusion. ([@k82cn](https://github.com/k82cn))
  * [#536](https://github.com/kubernetes-sigs/kube-batch/pull/536) Migrate from Godep to dep ([@Jeffwan](https://github.com/Jeffwan))
  * [#533](https://github.com/kubernetes-sigs/kube-batch/pull/533) The design doc of PodGroup Phase in Status. ([@k82cn](https://github.com/k82cn))
  * [#540](https://github.com/kubernetes-sigs/kube-batch/pull/540) Examine all pending tasks in a job ([@Jeffwan](https://github.com/Jeffwan))
  * [#544](https://github.com/kubernetes-sigs/kube-batch/pull/544) The design doc of "Dynamic Plugin Configuration" ([@k82cn](https://github.com/k82cn))
  * [#525](https://github.com/kubernetes-sigs/kube-batch/pull/525) Added Phase and Conditions to PodGroup.Status struct ([@Zyqsempai](https://github.com/Zyqsempai))
  * [#547](https://github.com/kubernetes-sigs/kube-batch/pull/547) Fixed status protobuf id. ([@k82cn](https://github.com/k82cn))
  * [#549](https://github.com/kubernetes-sigs/kube-batch/pull/549) Added name when register plugin. ([@k82cn](https://github.com/k82cn))
  * [#550](https://github.com/kubernetes-sigs/kube-batch/pull/550) Added SchedulerConfiguration type. ([@k82cn](https://github.com/k82cn))
  * [#551](https://github.com/kubernetes-sigs/kube-batch/pull/551) Upgrade k8s dependencies to v1.13 ([@Jeffwan](https://github.com/Jeffwan))
  * [#535](https://github.com/kubernetes-sigs/kube-batch/pull/535) Add Unschedulable PodCondition for pods in pending ([@Jeffwan](https://github.com/Jeffwan))
  * [#552](https://github.com/kubernetes-sigs/kube-batch/pull/552) Multi tiers for plugins. ([@k82cn](https://github.com/k82cn))
  * [#548](https://github.com/kubernetes-sigs/kube-batch/pull/548) Add e2e test for different task resource requests ([@Jeffwan](https://github.com/Jeffwan))
  * [#556](https://github.com/kubernetes-sigs/kube-batch/pull/556) Added VolumeScheduling. ([@k82cn](https://github.com/k82cn))
  * [#558](https://github.com/kubernetes-sigs/kube-batch/pull/558) Reduce verbosity level for recurring logs ([@mateusz-ciesielski](https://github.com/mateusz-ciesielski))
  * [#560](https://github.com/kubernetes-sigs/kube-batch/pull/560) Update PodGroup status. ([@k82cn](https://github.com/k82cn))
  * [#565](https://github.com/kubernetes-sigs/kube-batch/pull/565) Removed reclaim&preempt by default. ([@k82cn](https://github.com/k82cn))

## v0.3

### Issues:

  * [#499](https://github.com/kubernetes-sigs/kube-batch/pull/499) Added Statement for eviction in batch. ([@k82cn](http://github.com/k82cn))
  * [#491](https://github.com/kubernetes-sigs/kube-batch/pull/491) Update nodeInfo once node was cordoned ([@jiaxuanzhou](http://github.com/jiaxuanzhou))
  * [#490](https://github.com/kubernetes-sigs/kube-batch/pull/490) Add version options for kube-batch ([@jiaxuanzhou](http://github.com/jiaxuanzhou))
  * [#489](https://github.com/kubernetes-sigs/kube-batch/pull/489) Added Statement. ([@k82cn](http://github.com/k82cn))
  * [#480](https://github.com/kubernetes-sigs/kube-batch/pull/480) Added Pipelined Pods as valid tasks. ([@k82cn](http://github.com/k82cn))
  * [#468](https://github.com/kubernetes-sigs/kube-batch/pull/468) Detailed 'unschedulable' events ([@adam-marek](http://github.com/adam-marek))
  * [#470](https://github.com/kubernetes-sigs/kube-batch/pull/470) Fixed event recorder schema to include definitions for non kube-batch entities ([@adam-marek](http://github.com/adam-marek))
  * [#466](https://github.com/kubernetes-sigs/kube-batch/pull/466) Ordered Job by creation time. ([@k82cn](http://github.com/k82cn))
  * [#465](https://github.com/kubernetes-sigs/kube-batch/pull/465) Enable PDB based gang scheduling with discrete queues. ([@adam-marek](http://github.com/adam-marek))
  * [#464](https://github.com/kubernetes-sigs/kube-batch/pull/464) Order gang scheduled jobs based on time of arrival ([@adam-marek](http://github.com/adam-marek))
  * [#455](https://github.com/kubernetes-sigs/kube-batch/pull/455) Check node spec for unschedulable as well ([@jeefy](http://github.com/jeefy))
  * [#443](https://github.com/kubernetes-sigs/kube-batch/pull/443) Checked pod group unschedulable event. ([@k82cn](http://github.com/k82cn))
  * [#442](https://github.com/kubernetes-sigs/kube-batch/pull/442) Checked allowed pod number on node. ([@k82cn](http://github.com/k82cn))
  * [#440](https://github.com/kubernetes-sigs/kube-batch/pull/440) Updated node info if taints and lables changed. ([@k82cn](http://github.com/k82cn))
  * [#438](https://github.com/kubernetes-sigs/kube-batch/pull/438) Using jobSpec for queue e2e. ([@k82cn](http://github.com/k82cn))
  * [#433](https://github.com/kubernetes-sigs/kube-batch/pull/433) Added backfill action for BestEffort pods. ([@k82cn](http://github.com/k82cn))
  * [#418](https://github.com/kubernetes-sigs/kube-batch/pull/418) Supported Pod Affinity/Anti-Affinity Predicate. ([@k82cn](http://github.com/k82cn))
  * [#417](https://github.com/kubernetes-sigs/kube-batch/pull/417) fix BestEffort for overused ([@chenyangxueHDU](http://github.com/chenyangxueHDU))
  * [#414](https://github.com/kubernetes-sigs/kube-batch/pull/414) Added --schedule-period ([@k82cn](http://github.com/k82cn))

## v0.2

### Features

| Name                             | Version      | Notes                                                    |
| -------------------------------- | ------------ | -------------------------------------------------------- |
| Gang-scheduling/Coscheduling     | Experimental | [Doc](https://github.com/kubernetes/community/pull/2337) |
| Preemption/Reclaim               | Experimental |                                                          |
| Task Priority within Job         | Experimental |                                                          |
| Queue                            | Experimental |                                                          |
| DRF for Job sharing within Queue | Experimental |                                                          |


### Docker Images

```shell
docker pull kubesigs/kube-batch:v0.2
```

