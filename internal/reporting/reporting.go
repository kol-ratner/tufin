package reporting

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/kol-ratner/tufin/pkg/k8s"
)

func Status(msgChan chan<- string, kubeconfigPath string, cli *k8s.Client) error {
	pods, err := cli.Pods("default")
	if err != nil {
		return err
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"NAME", "READY", "STATUS", "RESTARTS", "START_TIME", "CPU_USAGE", "MEMORY_USAGE"})
	for _, pod := range pods.Items {
		util, err := cli.CalculateResourceUtilization(pod)
		if err != nil {
			return err
		}

		t.AppendRow(table.Row{
			pod.Name,
			pod.Status.ContainerStatuses[0].Ready,
			pod.Status.Phase,
			pod.Status.ContainerStatuses[0].RestartCount,
			pod.Status.StartTime.String(),
			util.CPU,
			util.Memory,
		})
		t.AppendSeparator()
	}

	t.Render()
	return nil
}
