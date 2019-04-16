/*
Copyright 2017 The Kubernetes Authors.

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

package scheduler

import (
	"time"

	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"

	schedcache "github.com/kubernetes-sigs/kube-batch/pkg/scheduler/cache"
	"github.com/kubernetes-sigs/kube-batch/pkg/scheduler/conf"
	"github.com/kubernetes-sigs/kube-batch/pkg/scheduler/framework"
	"github.com/kubernetes-sigs/kube-batch/pkg/scheduler/metrics"
)

type Scheduler struct {
	cache               schedcache.Cache
	config              *rest.Config
	actions             []framework.Action
	plugins             []conf.Tier
	schedulerConf       string
	schedulePeriod      time.Duration
	enableBackfill      bool
	starvationThreshold time.Duration
}

func NewScheduler(
	config *rest.Config,
	schedulerName string,
	conf string,
	period time.Duration,
	defaultQueue string,
) (*Scheduler, error) {
	scheduler := &Scheduler{
		config:         config,
		schedulerConf:  conf,
		cache:          schedcache.New(config, schedulerName, defaultQueue),
		schedulePeriod: period,
	}

	return scheduler, nil
}

func (pc *Scheduler) Run(stopCh <-chan struct{}) {
	var err error

	// Start cache for policy.
	go pc.cache.Run(stopCh)
	pc.cache.WaitForCacheSync(stopCh)

	// Load configuration of scheduler
	schedConf := defaultSchedulerConf
	if len(pc.schedulerConf) != 0 {
		if schedConf, err = readFile(pc.schedulerConf); err != nil {
			glog.Errorf("Failed to read scheduler configuration '%s', using default configuration: %v",
				pc.schedulerConf, err)
			schedConf = defaultSchedulerConf
		}
	}

	var config *conf.SchedulerConfiguration
	if config, err = loadConf(schedConf); err == nil {
		pc.plugins = config.Tiers
		pc.starvationThreshold = config.StarvationThreshold
		pc.enableBackfill = config.EnableBackfill
		pc.actions, err = getActions(config)
	}

	if err != nil {
		glog.Errorf("Failed to read scheduler configuration '%s': %s",
			schedConf, err)

		panic(err)
	}

	glog.V(4).Infof("pc setting %+v", pc)
	go wait.Until(pc.runOnce, pc.schedulePeriod, stopCh)
}

func (pc *Scheduler) runOnce() {
	glog.V(4).Infof("Start scheduling ...")
	scheduleStartTime := time.Now()
	defer glog.V(4).Infof("End scheduling ...")
	defer metrics.UpdateE2eDuration(metrics.Duration(scheduleStartTime))

	ssn := framework.OpenSession(pc.cache, pc.plugins)
	ssn.EnableBackfill = pc.enableBackfill
	ssn.StarvationThreshold = pc.starvationThreshold

	defer framework.CloseSession(ssn)

	glog.V(4).Infof("Start executing ...")
	for _, action := range pc.actions {
		actionStartTime := time.Now()
		action.Execute(ssn)
		metrics.UpdateActionDuration(action.Name(), metrics.Duration(actionStartTime))
	}
}
