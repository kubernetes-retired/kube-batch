package job

import (
	"fmt"
	"io"
	"os"

	"github.com/kubernetes-sigs/kube-batch/pkg/apis/scheduling/v1alpha1"
	"github.com/kubernetes-sigs/kube-batch/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// Weight of the queue
	Weight string = "Weight"

	// Default queue name
	Default string = "default"
)

type queueFlags struct {
	commonFlags

	Namespace string
	JobName   string
}

var queueJobFlags = &queueFlags{}

// InitQueueFlags is used to init all queue flags
func InitQueueFlags(cmd *cobra.Command) {
	initFlags(cmd, &queueJobFlags.commonFlags)

	cmd.Flags().StringVarP(&queueJobFlags.Namespace, "namespace", "", "default", "the namespace of job")
	cmd.Flags().StringVarP(&queueJobFlags.JobName, "name", "n", "", "the name of job")
}

// QueueJob creates commands to queue job
func JobQueue() error {
	queueName := Default

	config, err := buildConfig(queueJobFlags.Master, queueJobFlags.Kubeconfig)
	if err != nil {
		return err
	}

	jobClient := versioned.NewForConfigOrDie(config)
	job, err := jobClient.BatchV1alpha1().Jobs(queueJobFlags.Namespace).Get(queueJobFlags.JobName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if len(job.Spec.Queue) != 0 {
		queueName = job.Spec.Queue
	}

	queue, err := jobClient.SchedulingV1alpha1().Queues().Get(queueName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if queue == nil {
		fmt.Printf("No resources found\n")
		return nil
	}

	PrintQueue(queue, os.Stdout)

	return nil
}

// PrintQueue prints queue information
func PrintQueue(queue *v1alpha1.Queue, writer io.Writer) {
	_, err := fmt.Fprintf(writer, "%-25s%-8s\n",
		Name, Weight)
	if err != nil {
		fmt.Printf("Failed to print queue command result: %s.\n", err)
	}

	_, err = fmt.Fprintf(writer, "%-25s%-8d\n",
		queue.Name, queue.Spec.Weight)
	if err != nil {
		fmt.Printf("Failed to print queue command result: %s.\n", err)
	}

}
