package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

// type client struct {
// 	ClientSet *kubernetes.Clientset
// 	Metrics   *metrics.Clientset
// }

type Client struct {
	kubernetes.Interface
	Metrics metrics.Interface
}

func NewClient(config *rest.Config) (*Client, error) {
	client := &Client{}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	client.Interface = clientSet

	metricsClientSet, err := metrics.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	client.Metrics = metricsClientSet

	return client, nil
}
