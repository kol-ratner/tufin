package app

import (
	"context"

	"k8s.io/client-go/kubernetes"
)

type KubernetesResource int

const (
	Deployment KubernetesResource = iota
	Service
	ConfigMap
	Secret
	PVC
)

type Application struct {
	Client    kubernetes.Interface
	Config    *ApplicationConfig
	Resources []KubernetesResource
}

func NewApplication(client kubernetes.Interface, config *ApplicationConfig, resources []KubernetesResource) *Application {
	return &Application{
		Client:    client,
		Config:    config,
		Resources: resources,
	}
}

func (a *Application) Deploy() error {
	ctx := context.Background()

	for _, obj := range a.Resources {
		switch obj {
		case Deployment:
			if err := a.deployment(ctx); err != nil {
				return err
			}
		case Service:
			if err := a.service(ctx); err != nil {
				return err
			}
		case Secret:
			if err := a.secret(ctx); err != nil {
				return err
			}
		case PVC:
			if err := a.pvc(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}
