package test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("XQueueJob Plugin E2E Test", func() {
	It("XQueueJob Client Create Test", func() {
		context := initTestContext()
		defer cleanupTestContext(context)
		rep := clusterSize(context, oneCPU)

		xqueueJob := createXQueueJob(context, "xqj-1", 2, rep, workerPriority, "busybox", oneCPU)
		err := waitXJobCreated(context, xqueueJob.Name)
		Expect(err).NotTo(HaveOccurred())
	})

	It("XQueueJob Client CreateListDelete Test", func() {
		context := initTestContext()
		defer cleanupTestContext(context)
		rep := clusterSize(context, oneCPU)

		xqueueJob := createXQueueJob(context, "xqj-1", 2, rep, workerPriority, "busybox", oneCPU)

		err := waitXJobCreated(context, xqueueJob.Name)
		Expect(err).NotTo(HaveOccurred())

		err = listXQueueJob(context, "xqj-1")
		Expect(err).NotTo(HaveOccurred())

		err = deleteXQueueJob(context, "xqj-1")
		Expect(err).NotTo(HaveOccurred())

	})

	It("XQueueJob Client CreateListDelete Multiple Jobs Test", func() {
		context := initTestContext()
		defer cleanupTestContext(context)
		rep := clusterSize(context, oneCPU)

		xqueueJob := createXQueueJob(context, "xqj-1", 2, rep, workerPriority, "busybox", oneCPU)
		err := waitXJobCreated(context, xqueueJob.Name)
		Expect(err).NotTo(HaveOccurred())

		xqueueJob2 := createXQueueJob(context, "xqj-2", 2, rep, workerPriority, "busybox", oneCPU)

		err = waitXJobCreated(context, xqueueJob2.Name)
		Expect(err).NotTo(HaveOccurred())

		err = listXQueueJob(context, "xqj-1")
		Expect(err).NotTo(HaveOccurred())

		err = listXQueueJob(context, "xqj-2")
		Expect(err).NotTo(HaveOccurred())

		err = listXQueueJobs(context, 2)
		Expect(err).NotTo(HaveOccurred())

		err = deleteXQueueJob(context, "xqj-1")
		Expect(err).NotTo(HaveOccurred())

		err = deleteXQueueJob(context, "xqj-2")
		Expect(err).NotTo(HaveOccurred())
	})


	It("XQueueJob Preemption", func() {
		context := initTestContext()
		defer cleanupTestContext(context)

		slot := oneCPU
		rep := clusterSize(context, slot)

		qj1 := createXQueueJob(context, "preemptee-qj", 1, rep, workerPriority, "nginx", slot)
		err := waitXJobReady(context, qj1.Name, int(rep/2))
		Expect(err).NotTo(HaveOccurred())

		qj2 := createXQueueJob(context, "preemptor-qj", 1, rep, masterPriority, "nginx", slot)
		err = waitXJobReady(context, qj2.Name, int(rep/2))
		Expect(err).NotTo(HaveOccurred())
	})

	It("PreemptionRestart", func() {
		context := initTestContext()
		defer cleanupTestContext(context)

		slot := oneCPU
		rep := clusterSize(context, slot)

		qj1 := createXQueueJob(context, "preemptee-qj", 1, rep, workerPriority, "nginx", slot)
		err := waitXJobReady(context, qj1.Name, int(rep/2))
		Expect(err).NotTo(HaveOccurred())

		qj2 := createXQueueJob(context, "preemptor-qj", 1, rep, masterPriority, "nginx", slot)

		err = waitXJobReady(context, qj2.Name, int(rep/2))
		Expect(err).NotTo(HaveOccurred())

		err = deleteXQueueJob(context, "preemptor-qj")
		Expect(err).NotTo(HaveOccurred())

		err = waitXJobReady(context, qj1.Name, int(rep/2))
		Expect(err).NotTo(HaveOccurred())
	})

})
