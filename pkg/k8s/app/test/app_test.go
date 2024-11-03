package app_test

import (
	"testing"

	"github.com/kol-ratner/tufin/pkg/k8s/app"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes/fake"
)

func TestApplication_Service(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()

	tests := []struct {
		name      string
		config    *app.ApplicationConfig
		wantError bool
	}{
		{
			name: "valid service config",
			config: &app.ApplicationConfig{
				Name:      "test-app",
				Namespace: "default",
				Labels: map[string]string{
					"app": "test-app",
				},
				Svc: app.SvcConfig{
					Port: 8080,
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			application := &app.Application{
				Client:    fakeClientset,
				Config:    tt.config,
				Resources: []app.KubernetesResource{app.Service},
			}
			err := application.Deploy()
			if (err != nil) != tt.wantError {
				t.Errorf("Service() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestApplication_Deployment(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()

	tests := []struct {
		name      string
		config    *app.ApplicationConfig
		wantError bool
	}{
		{
			name: "basic deployment",
			config: &app.ApplicationConfig{
				Name:      "test-app",
				Namespace: "default",
				Labels: map[string]string{
					"app": "test-app",
				},
				Deployment: app.DeploymentConfig{
					Replicas:      3,
					Image:         "nginx:latest",
					ContainerPort: 80,
					SelectorMatchLabels: map[string]string{
						"app": "test-app",
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("256Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("512Mi"),
						},
					},
				},
			},
			wantError: false,
		},
		{
			name: "deployment with volumes and env vars",
			config: &app.ApplicationConfig{
				Name:      "test-app-storage",
				Namespace: "default",
				Labels: map[string]string{
					"app": "test-app-storage",
				},
				Deployment: app.DeploymentConfig{
					Replicas:      1,
					Image:         "mysql:5.7",
					ContainerPort: 3306,
					EnvVars: []corev1.EnvVar{
						{
							Name:  "MYSQL_ROOT_PASSWORD",
							Value: "test-password",
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "mysql-storage",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "mysql-pvc",
								},
							},
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "mysql-storage",
							MountPath: "/var/lib/mysql",
						},
					},
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			application := &app.Application{
				Client:    fakeClientset,
				Config:    tt.config,
				Resources: []app.KubernetesResource{app.Deployment},
			}
			err := application.Deploy()
			if (err != nil) != tt.wantError {
				t.Errorf("Deployment() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestApplication_Secret(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()

	tests := []struct {
		name      string
		config    *app.ApplicationConfig
		wantError bool
	}{
		{
			name: "basic secret",
			config: &app.ApplicationConfig{
				Name:      "test-app",
				Namespace: "default",
				Labels: map[string]string{
					"app": "test-app",
				},
				Secret: app.SecretConfig{
					SecretName: "mysql-creds",
					SecretType: "Opaque",
					SecretData: map[string][]byte{
						"username": []byte("admin"),
						"password": []byte("secret123"),
					},
				},
			},
			wantError: false,
		},
		{
			name: "secret with multiple keys",
			config: &app.ApplicationConfig{
				Name:      "mysql-creds",
				Namespace: "default",
				Labels: map[string]string{
					"app": "mysql",
				},
				Secret: app.SecretConfig{
					SecretName: "mysql-creds",
					SecretType: "Opaque",
					SecretData: map[string][]byte{
						"MYSQL_ROOT_PASSWORD": []byte("rootpass"),
						"MYSQL_PASSWORD":      []byte("userpass"),
						"MYSQL_USER":          []byte("wordpress"),
					},
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			application := &app.Application{
				Client:    fakeClientset,
				Config:    tt.config,
				Resources: []app.KubernetesResource{app.Secret},
			}
			err := application.Deploy()
			if (err != nil) != tt.wantError {
				t.Errorf("Secret() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestApplication_PVC(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()

	tests := []struct {
		name      string
		config    *app.ApplicationConfig
		wantError bool
	}{
		{
			name: "basic pvc",
			config: &app.ApplicationConfig{
				Name:      "test-storage",
				Namespace: "default",
				Labels: map[string]string{
					"app": "test-app",
				},
				Pvc: app.PvcConfig{
					AccessMode: corev1.ReadWriteOnce,
					Size:       "5Gi",
				},
			},
			wantError: false,
		},
		{
			name: "pvc with custom storage class",
			config: &app.ApplicationConfig{
				Name:      "mysql-storage",
				Namespace: "default",
				Labels: map[string]string{
					"app": "mysql",
				},
				Pvc: app.PvcConfig{
					AccessMode: corev1.ReadOnlyMany,
					Size:       "10Gi",
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			application := &app.Application{
				Client:    fakeClientset,
				Config:    tt.config,
				Resources: []app.KubernetesResource{app.PVC},
			}
			err := application.Deploy()
			if (err != nil) != tt.wantError {
				t.Errorf("PVC() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
