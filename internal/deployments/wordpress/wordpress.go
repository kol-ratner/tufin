package wordpress

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"

	"github.com/kol-ratner/tufin/internal/config"
	k8sapp "github.com/kol-ratner/tufin/pkg/k8s/app"
)

func New(cliSet kubernetes.Interface, opts ...config.Option) k8sapp.Application {

	cfg := newConfig(opts...)

	return k8sapp.Application{
		Client: cliSet,
		Config: cfg,
		Resources: []k8sapp.KubernetesResource{
			k8sapp.Deployment,
			k8sapp.PVC,
			k8sapp.Service,
		},
	}
}

func newConfig(opts ...config.Option) *k8sapp.ApplicationConfig {
	name := "wordpress"

	cfg := &k8sapp.ApplicationConfig{
		Name:      name,
		Namespace: "default",
		Labels: map[string]string{
			"app":  name,
			"tier": "frontend",
		},

		Deployment: k8sapp.DeploymentConfig{
			Replicas: 1,
			Image:    "wordpress:6.2.1-apache",
			SelectorMatchLabels: map[string]string{
				"app":  name,
				"tier": "frontend",
			},
			ContainerPort: 80,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("250m"),
					corev1.ResourceMemory: resource.MustParse("256Mi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("512Mi"),
				},
			},
			EnvVars: []corev1.EnvVar{
				{
					Name:  "WORDPRESS_DB_HOST",
					Value: "mysql",
				},
				{
					Name:  "WORDPRESS_DB_USER",
					Value: "wordpress",
				},
				{
					Name: "WORDPRESS_DB_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "mysql-creds",
							},
							Key: "password",
						},
					},
				},
			},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      name,
					MountPath: "/var/www/html",
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: name,
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
			Size:       "2Gi",
		},

		Svc: k8sapp.SvcConfig{
			Port:             80,
			DisableClusterIP: false,
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
