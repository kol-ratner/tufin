package reporting

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/kol-ratner/tufin/pkg/k8s"
)

func Status(msgChan chan<- string, cli *k8s.Client) error {
	pods, err := cli.Pods("default")
	if err != nil {
		return err
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"NAME", "READY", "STATUS", "RESTARTS", "START_TIME", "CPU_USAGE", "MEMORY_USAGE"})
	for _, pod := range pods.Items {
		var cpuUsage, memoryUsage, startTime string
		var ready bool
		var restartCount int32

		if pod.Status.Phase == "Running" {
			if util, err := cli.CalculateResourceUtilization(pod); err == nil {
				cpuUsage = util.CPU
				memoryUsage = util.Memory
			}
		}

		if pod.Status.StartTime != nil {
			startTime = pod.Status.StartTime.String()
		}

		// Safely access container statuses
		if len(pod.Status.ContainerStatuses) > 0 {
			ready = pod.Status.ContainerStatuses[0].Ready
			restartCount = pod.Status.ContainerStatuses[0].RestartCount
		}

		t.AppendRow(table.Row{
			pod.Name,
			ready,
			pod.Status.Phase,
			restartCount,
			startTime,
			cpuUsage,
			memoryUsage,
		})
		t.AppendSeparator()
	}

	t.Render()
	return nil
}
