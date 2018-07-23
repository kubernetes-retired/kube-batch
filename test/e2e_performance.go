package test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"fmt"
	"time"
)

var _ = Describe("XQueueJob Plugin E2E Test", func() {
        It("GangScheduling", func() {
		context := initTestContext()
                defer cleanupTestContext(context)
		slot := oneCPU
		ctime := time.Now().Unix()
		// create 100 jobs and measure end-to-end latency and avg/std in wait time for each job
		for i := 0; i < 20; i++ {
			name := fmt.Sprintf("qj-%v", i)
			createXQueueJob(context, name, 2, 2, workerPriority, "busybox", slot)
		}
		err := listXQueueJobs(context, 0)
		Expect(err).NotTo(HaveOccurred())

		// wait all jobs to finish
		ftime := time.Now().Unix()

		diff := ftime - ctime
		fmt.Printf("Makespan of workload is %v\n", diff)
	})
})
