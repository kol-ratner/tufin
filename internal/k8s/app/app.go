package app

import (
	"context"

	corev1 "k8s.io/api/core/v1"
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

type SvcConfig struct {
	Port             int32
	DisableClusterIP bool
}

type ApplicationConfig struct {
	Name      string
	Labels    map[string]string
	Namespace string

	Image         string
	ContainerPort int32
	Replicas      int32
	Resources     corev1.ResourceRequirements

	Volumes      []corev1.Volume
	EnvVars      []corev1.EnvVar
	VolumeMounts []corev1.VolumeMount

	PersistentVolumeSize string

	Svc SvcConfig

	SecretType corev1.SecretType
	SecretData map[string][]byte
}

type Application struct {
	Client    *kubernetes.Clientset
	Config    *ApplicationConfig
	Resources []KubernetesResource
}

func NewApplication(client *kubernetes.Clientset, config *ApplicationConfig, resources []KubernetesResource) *Application {
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
