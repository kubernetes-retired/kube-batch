package scheduler

import (
	"github.com/kubernetes-sigs/kube-batch/pkg/scheduler/conf"
	"reflect"
	"testing"
	"time"
)

func TestAddNode(t *testing.T) {
	var conf1 = `
starvation-threshold: 20h
enable-backfill: true
actions: "allocate, backfill"
tiers:
- plugins:
  - name: priority
  - name: gang
- plugins:
  - name: drf
  - name: predicates
`
	tests := []struct {
		name      string
		configStr string
		expected  *conf.SchedulerConfiguration
	}{
		{
			name:      "configuration with starvation-threshold",
			configStr: conf1,
			expected: &conf.SchedulerConfiguration{
				StarvationThreshold: 20 * time.Hour,
				EnableBackfill:      true,
				Actions:             "allocate, backfill",
				Tiers: []conf.Tier{
					{
						Plugins: []conf.PluginOption{
							{
								Name: "priority",
							},
							{
								Name: "gang",
							},
						},
					},
					{
						Plugins: []conf.PluginOption{
							{
								Name: "drf",
							},
							{
								Name: "predicates",
							},
						},
					},
				},
			},
		},
		{
			name:      "empty configuration with default starvation-threshold",
			configStr: "",
			expected: &conf.SchedulerConfiguration{
				EnableBackfill:      false,
				StarvationThreshold: conf.DefaultStarvingThreshold,
			},
		},
	}

	for i, test := range tests {
		actual, _ := loadConf(test.configStr)

		if !reflect.DeepEqual(test.expected, actual) {
			t.Errorf("case(%d) %s: \n expected %v, \n got %v \n",
				i, test.name, test.expected, actual)
		}
	}
}
