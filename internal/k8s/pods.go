package k8s

import (
	"context"
	"errors"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func (c *client) Pods(namspace string) (*v1.PodList, error) {
	pods, err := c.ClientSet.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return pods, nil
}

type resources struct {
	Requests v1.ResourceList
	Limits   v1.ResourceList
}

func (c *client) PodResources(pod v1.Pod) *resources {
	r := &resources{}

	r.Limits = pod.Spec.Containers[0].Resources.Limits
	r.Requests = pod.Spec.Containers[0].Resources.Requests

	return r
}

func (c *client) PodMetrics(pod v1.Pod) (*v1beta1.PodMetrics, error) {
	metrics, err := c.Metrics.MetricsV1beta1().
		PodMetricses(pod.Namespace).
		Get(context.Background(), pod.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

type podUtilization struct {
	CPU    string
	Memory string
}

func (c *client) CalculateResourceUtilization(pod v1.Pod) (podUtilization, error) {
	cpuReq := c.PodResources(pod).Requests.Cpu()
	memoryReq := c.PodResources(pod).Requests.Memory()

	metrics, err := c.PodMetrics(pod)
	if err != nil {
		return podUtilization{}, err
	}

	cpuUtil := metrics.Containers[0].Usage.Cpu()
	memUtil := metrics.Containers[0].Usage.Memory()

	if memoryReq.IsZero() {
		return podUtilization{}, errors.New("pod memory request not configured")
	} else if cpuReq.IsZero() {
		return podUtilization{}, errors.New("pod cpu request not configured")
	}

	cpuPercentage := float64(cpuUtil.MilliValue()) / float64(cpuReq.MilliValue()) * 100
	memPercentage := float64(memUtil.Value()) / float64(memoryReq.Value()) * 100

	return podUtilization{
		CPU:    formatPercentage(cpuPercentage),
		Memory: formatPercentage(memPercentage),
	}, nil
}

func formatPercentage(percentage float64) string {
	return fmt.Sprintf("%.1f%%", percentage)
}
