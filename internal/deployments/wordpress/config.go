package wordpress

import (
	k8sapp "github.com/kol-ratner/tufin/internal/k8s/app"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type ConfigOverrides struct {
	Replicas      int32
	CPURequest    string
	MemoryRequest string
	CPULimit      string
	MemoryLimit   string
	VolumeSize    string
}

func WithReplicas(replicas int32) Option {
	return func(co *ConfigOverrides) {
		co.Replicas = replicas
	}
}

func WithCPURequest(cpu string) Option {
	return func(co *ConfigOverrides) {
		co.CPURequest = cpu
	}
}

func WithMemoryRequest(mem string) Option {
	return func(co *ConfigOverrides) {
		co.MemoryRequest = mem
	}
}

func WithCPULimit(cpu string) Option {
	return func(co *ConfigOverrides) {
		co.CPULimit = cpu
	}
}

func WithMemoryLimit(mem string) Option {
	return func(co *ConfigOverrides) {
		co.MemoryLimit = mem
	}
}

func WithVolumeSize(size string) Option {
	return func(co *ConfigOverrides) {
		co.VolumeSize = size
	}
}

func newConfig(opts ...Option) *k8sapp.ApplicationConfig {
	name := "wordpress"

	cfg := &k8sapp.ApplicationConfig{
		Name:      name,
		Namespace: "default",
		Image:     "wordpress:6.2.1-apache",
		Labels: map[string]string{
			"app":                    name,
			"app.kubernetes.io/name": name,
		},
		ContainerPort: 80,
		Replicas:      1,
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
		PersistentVolumeSize: "2Gi",
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
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      name,
				MountPath: "/var/www/html",
			},
		},
		Svc: k8sapp.SvcConfig{
			Port:             80,
			DisableClusterIP: false,
		},
	}

	// Create and apply overrides
	overrides := &ConfigOverrides{}
	for _, opt := range opts {
		opt(overrides)
	}

	// Apply overrides to the config
	if overrides.Replicas != 0 {
		cfg.Replicas = overrides.Replicas
	}
	if overrides.CPURequest != "" {
		cfg.Resources.Requests[corev1.ResourceCPU] = resource.MustParse(overrides.CPURequest)
	}
	if overrides.MemoryRequest != "" {
		cfg.Resources.Requests[corev1.ResourceMemory] = resource.MustParse(overrides.MemoryRequest)
	}
	if overrides.CPULimit != "" {
		cfg.Resources.Limits[corev1.ResourceCPU] = resource.MustParse(overrides.CPULimit)
	}
	if overrides.MemoryLimit != "" {
		cfg.Resources.Limits[corev1.ResourceMemory] = resource.MustParse(overrides.MemoryLimit)
	}
	if overrides.VolumeSize != "" {
		cfg.PersistentVolumeSize = overrides.VolumeSize
	}

	return cfg
}
