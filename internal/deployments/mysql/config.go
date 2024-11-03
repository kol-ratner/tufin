package mysql

import (
	"fmt"

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
}

type Option func(*ConfigOverrides)

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

func newConfig(opts ...Option) *k8sapp.ApplicationConfig {
	name := "mysql"

	cfg := &k8sapp.ApplicationConfig{
		Name:      name,
		Namespace: "default",
		Image:     "mysql:8.0",
		Labels: map[string]string{
			"app":                    name,
			"app.kubernetes.io/name": name,
		},
		ContainerPort: 3306,
		Replicas:      1,
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
		PersistentVolumeSize: "5Gi",
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
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      fmt.Sprintf("%-storage", name),
				MountPath: "/var/lib/mysql",
			},
		},
		Svc: k8sapp.SvcConfig{
			Port:             3306,
			DisableClusterIP: true,
		},
		SecretType: "Opaque",
		SecretData: map[string][]byte{
			"password": k8sapp.GeneratePassword(25),
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

	return cfg
}
