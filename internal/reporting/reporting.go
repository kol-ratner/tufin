package reporting

import (
	"context"
	"fmt"
	"os"

	"github.com/kol-ratner/tufin/internal/k8s"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/jedib0t/go-pretty/v6/table"
)

func Status(msgChan chan<- string) error {
	// grabbing the kubeconfig from the hosts default ~/.kube/config file
	kubeConfig, err := k8s.GetKubeConfigFromHost("")
	if err != nil {
		return err
	}

	stdClientSet, err := k8s.NewClient(kubeConfig)
	if err != nil {
		return err
	}
	metricsClientSet, err := metrics.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}

	pods, err := stdClientSet.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"NAME", "READY", "STATUS", "RESTARTS", "START_TIME", "CPU_USAGE", "MEMORY_USAGE"})

	for _, pod := range pods.Items {
		cpuLimit := pod.Spec.Containers[0].Resources.Limits.Cpu()
		memLimit := pod.Spec.Containers[0].Resources.Limits.Memory()

		fmt.Println(cpuLimit)
		fmt.Println(memLimit)

		metrics, err := metricsClientSet.MetricsV1beta1().PodMetricses("default").Get(context.Background(), pod.Name, metav1.GetOptions{})
		if err != nil {
			continue
		}

		cpuUsage := calculateResourcePercentage(metrics.Containers[0].Usage.Cpu(), cpuLimit)
		memUsage := calculateResourcePercentage(metrics.Containers[0].Usage.Memory(), memLimit)

		t.AppendRow(table.Row{
			pod.Name,
			pod.Status.ContainerStatuses[0].Ready,
			pod.Status.Phase,
			pod.Status.ContainerStatuses[0].RestartCount,
			pod.Status.StartTime.String(),
			cpuUsage,
			memUsage,
		})
		t.AppendSeparator()
	}

	t.Render()

	return nil
}

func calculateResourcePercentage(usage, limit *resource.Quantity) string {
	if limit.IsZero() {
		return "No Limit"
	}
	percentage := float64(usage.MilliValue()) / float64(limit.MilliValue()) * 100
	return fmt.Sprintf("%.1f%%", percentage)
}
