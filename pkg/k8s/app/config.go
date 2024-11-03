package app

import (
	corev1 "k8s.io/api/core/v1"
)

type DeploymentConfig struct {
	Replicas            int32
	Image               string
	ContainerPort       int32
	SelectorMatchLabels map[string]string
	Resources           corev1.ResourceRequirements

	EnvVars      []corev1.EnvVar
	Volumes      []corev1.Volume
	VolumeMounts []corev1.VolumeMount
}

type PvcConfig struct {
	AccessMode corev1.PersistentVolumeAccessMode
	Size       string
}

type SvcConfig struct {
	Port             int32
	DisableClusterIP bool
}

type SecretConfig struct {
	SecretName string
	SecretType corev1.SecretType
	SecretData map[string][]byte
}

type ApplicationConfig struct {
	Name      string
	Labels    map[string]string
	Namespace string

	Pvc        PvcConfig
	Deployment DeploymentConfig
	Svc        SvcConfig
	Secret     SecretConfig
}
