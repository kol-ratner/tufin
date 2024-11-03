package mysql

import (
	k8sapp "github.com/kol-ratner/tufin/internal/k8s/app"
	"k8s.io/client-go/kubernetes"
)

func New(cliSet *kubernetes.Clientset, opts ...Option) k8sapp.Application {

	cfg := newConfig(opts...)

	return k8sapp.Application{
		Client: cliSet,
		Config: cfg,
		Resources: []k8sapp.KubernetesResource{
			k8sapp.Deployment,
			k8sapp.PVC,
			k8sapp.Service,
			k8sapp.Secret,
		},
	}
}
