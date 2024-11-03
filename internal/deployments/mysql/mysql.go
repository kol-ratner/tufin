package mysql

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"

	"github.com/kol-ratner/tufin/internal/config"
	k8sapp "github.com/kol-ratner/tufin/pkg/k8s/app"
)

func New(cliSet *kubernetes.Clientset, opts ...config.Option) k8sapp.Application {

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

func newConfig(opts ...config.Option) *k8sapp.ApplicationConfig {
	name := "mysql"

	cfg := &k8sapp.ApplicationConfig{
		Name:      name,
		Namespace: "default",
		Labels: map[string]string{
			"app":                    name,
			"app.kubernetes.io/name": name,
		},

		Deployment: k8sapp.DeploymentConfig{
			Replicas: 1,
			Image:    "mysql:8.0",
			SelectorMatchLabels: map[string]string{
				"app":                    name,
				"app.kubernetes.io/name": name,
			},
			ContainerPort: 3306,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("750Mi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
			},
			EnvVars: []corev1.EnvVar{
				{
					Name: "MYSQL_ROOT_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: fmt.Sprintf("%s-creds", name),
							},
							Key: "password",
						},
					},
				},
				{
					Name:  "MYSQL_DATABASE",
					Value: "wordpress",
				},
				{
					Name:  "MYSQL_USER",
					Value: "wordpress",
				},
				{
					Name: "MYSQL_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: fmt.Sprintf("%s-creds", name),
							},
							Key: "password",
						},
					},
				},
			},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      fmt.Sprintf("%-storage", name),
					MountPath: "/var/lib/mysql",
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: fmt.Sprintf("%-storage", name),
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: name,
						},
					},
				},
			},
		},

		Pvc: k8sapp.PvcConfig{
			AccessMode: corev1.ReadWriteOnce,
			Size:       "5Gi",
		},

		Svc: k8sapp.SvcConfig{
			Port:             3306,
			DisableClusterIP: true,
		},

		Secret: k8sapp.SecretConfig{
			SecretName: fmt.Sprintf("%s-creds", name),
			SecretType: "Opaque",
			SecretData: map[string][]byte{
				"password": k8sapp.GeneratePassword(25),
			},
		},
	}

	// Create and apply overrides
	overrides := &config.DeploymentOverrides{}
	for _, opt := range opts {
		opt(overrides)
	}

	// Apply overrides to the config
	if overrides.Replicas != 0 {
		cfg.Deployment.Replicas = overrides.Replicas
	}
	if overrides.CPURequest != "" {
		cfg.Deployment.Resources.Requests[corev1.ResourceCPU] = resource.MustParse(overrides.CPURequest)
	}
	if overrides.MemoryRequest != "" {
		cfg.Deployment.Resources.Requests[corev1.ResourceMemory] = resource.MustParse(overrides.MemoryRequest)
	}
	if overrides.CPULimit != "" {
		cfg.Deployment.Resources.Limits[corev1.ResourceCPU] = resource.MustParse(overrides.CPULimit)
	}
	if overrides.MemoryLimit != "" {
		cfg.Deployment.Resources.Limits[corev1.ResourceMemory] = resource.MustParse(overrides.MemoryLimit)
	}
	if overrides.VolumeSize != "" {
		cfg.Pvc.Size = overrides.VolumeSize
	}

	return cfg
}
