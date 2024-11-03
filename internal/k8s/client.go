package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

type client struct {
	ClientSet *kubernetes.Clientset
	Metrics   *metrics.Clientset
}

func NewClient(config *rest.Config) (*client, error) {
	client := &client{}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	client.ClientSet = clientSet

	metricsClientSet, err := metrics.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	client.Metrics = metricsClientSet

	return client, nil
}
