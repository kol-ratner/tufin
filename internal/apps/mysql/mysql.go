package mysql

import (
	"context"
	"fmt"

	"golang.org/x/exp/rand"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Options struct {
	Replicas      int32
	CPURequest    string
	MemoryRequest string
	CPULimit      string
	MemoryLimit   string
}

type mysql struct {
	cliSet *kubernetes.Clientset

	namespace string
	name      string
	replicas  int32
	image     string
	port      int32
	resources corev1.ResourceRequirements
}

func New(cliSet *kubernetes.Clientset, opts *Options) *mysql {
	defaultOpts := Options{
		Replicas:      1,
		CPURequest:    "1",
		MemoryRequest: "1Gi",
		CPULimit:      "1",
		MemoryLimit:   "1Gi",
	}

	if opts == nil {
		opts = &defaultOpts
	}

	if opts != nil {
		if opts.Replicas == 0 {
			opts.Replicas = defaultOpts.Replicas
		}
		if opts.CPURequest == "" {
			opts.CPURequest = defaultOpts.CPURequest
		}
		if opts.CPULimit == "" {
			opts.CPULimit = defaultOpts.CPULimit
		}
		if opts.MemoryRequest == "" {
			opts.MemoryRequest = defaultOpts.MemoryRequest
		}
		if opts.MemoryLimit == "" {
			opts.MemoryLimit = defaultOpts.MemoryLimit
		}
	}

	return &mysql{
		cliSet:    cliSet,
		namespace: "default",
		name:      "mysql",
		replicas:  opts.Replicas,
		image:     "mysql:8.0",
		port:      3306,
		resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(opts.CPULimit),
				corev1.ResourceMemory: resource.MustParse(opts.MemoryLimit),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(opts.CPURequest),
				corev1.ResourceMemory: resource.MustParse(opts.MemoryRequest),
			},
		},
	}
}

func (m *mysql) Deploy() error {
	ctx := context.Background()

	if err := m.secret(ctx); err != nil {
		return err
	}

	if err := m.pvc(ctx); err != nil {
		return err
	}

	if err := m.deployment(ctx); err != nil {
		return err
	}

	if err := m.service(ctx); err != nil {
		return err
	}
	return nil
}

func generatePassword(length int) []byte {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	pass := make([]byte, length)
	for i := range pass {
		pass[i] = charset[rand.Intn(len(charset))]
	}
	return pass
}

func (m *mysql) secret(ctx context.Context) error {
	scrtCli := m.cliSet.CoreV1().Secrets(m.namespace)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-creds", m.name),
			Namespace: m.namespace,
			Labels: map[string]string{
				"app": m.name,
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"password": generatePassword(25),
		},
	}

	_, err := scrtCli.Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			scrtCli.Update(ctx, secret, metav1.UpdateOptions{})
			return nil
		}
		return err
	}
	return nil
}

func (m *mysql) service(ctx context.Context) error {
	svcCli := m.cliSet.CoreV1().Services(m.namespace)

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.name,
			Namespace: m.namespace,
			Labels: map[string]string{
				"app":                    m.name,
				"app.kubernetes.io/name": m.name,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":  m.name,
				"tier": m.name,
			},
			ClusterIP: "None",
			Ports: []corev1.ServicePort{
				{
					Name:     m.name,
					Protocol: corev1.ProtocolTCP,
					Port:     m.port,
				},
			},
		},
	}

	_, err := svcCli.Create(ctx, svc, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			svcCli.Update(ctx, svc, metav1.UpdateOptions{})
			return nil
		}
		return err
	}

	return nil
}

func (m *mysql) pvc(ctx context.Context) error {
	pvcCli := m.cliSet.CoreV1().PersistentVolumeClaims(m.namespace)

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.name,
			Namespace: m.namespace,
			Labels: map[string]string{
				"app": m.name,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("5Gi"),
				},
			},
		},
	}

	_, err := pvcCli.Create(ctx, pvc, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			pvcCli.Update(ctx, pvc, metav1.UpdateOptions{})
			return nil
		}
		return err
	}

	return nil
}

func (m *mysql) deployment(ctx context.Context) error {
	dCli := m.cliSet.AppsV1().Deployments(m.namespace)

	deployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.name,
			Namespace: m.namespace,
			Labels: map[string]string{
				"app": m.name,
			},
		},
		Spec: v1.DeploymentSpec{
			Replicas: &m.replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  m.name,
					"tier": m.name,
				},
			},
			Strategy: v1.DeploymentStrategy{
				Type: v1.RecreateDeploymentStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  m.name,
						"tier": m.name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  m.name,
							Image: m.image,
							Env: []corev1.EnvVar{
								{
									Name: "MYSQL_ROOT_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: fmt.Sprintf("%s-creds", m.name),
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
												Name: fmt.Sprintf("%s-creds", m.name),
											},
											Key: "password",
										},
									},
								},
							},
							Resources: m.resources,
							Ports: []corev1.ContainerPort{
								{
									Name:          m.name,
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: m.port,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      fmt.Sprintf("%-storage", m.name),
									MountPath: "/var/lib/mysql",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: fmt.Sprintf("%-storage", m.name),
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: m.name,
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := dCli.Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			dCli.Update(ctx, deployment, metav1.UpdateOptions{})
			return nil
		}
		return err

	}

	return nil

}
