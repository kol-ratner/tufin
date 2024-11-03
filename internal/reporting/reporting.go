package reporting

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/kol-ratner/tufin/pkg/k8s"
)

func Status(msgChan chan<- string) error {
	// grabbing the kubeconfig from the hosts default ~/.kube/config file
	kubeConfig, err := k8s.GetKubeConfigFromHost("")
	if err != nil {
		return err
	}

	k8s, err := k8s.NewClient(kubeConfig)
	if err != nil {
		return err
	}

	pods, err := k8s.Pods("default")
	if err != nil {
		return err
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"NAME", "READY", "STATUS", "RESTARTS", "START_TIME", "CPU_USAGE", "MEMORY_USAGE"})
	for _, pod := range pods.Items {
		util, err := k8s.CalculateResourceUtilization(pod)
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
