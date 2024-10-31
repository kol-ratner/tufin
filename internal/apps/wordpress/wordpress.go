package wordpress

import (
	"context"

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
	WWWVolumeSize string
}

type wordpress struct {
	cliSet *kubernetes.Clientset

	namespace     string
	name          string
	replicas      int32
	image         string
	port          int32
	resources     corev1.ResourceRequirements
	wwwVolumeSize string
}

func New(cliSet *kubernetes.Clientset, opts *Options) *wordpress {
	defaultOpts := Options{
		Replicas:      1,
		CPURequest:    "500m",
		MemoryRequest: "512Mi",
		CPULimit:      "1",
		MemoryLimit:   "1Gi",
		WWWVolumeSize: "5Gi",
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
		if opts.WWWVolumeSize == "" {
			opts.WWWVolumeSize = defaultOpts.WWWVolumeSize
		}
	}

	return &wordpress{
		cliSet:    cliSet,
		namespace: "default",
		name:      "wordpress",
		replicas:  1,
		image:     "wordpress:6.2.1-apache",
		port:      80,
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
		wwwVolumeSize: opts.WWWVolumeSize,
	}
}

func (w *wordpress) Deploy() error {
	ctx := context.Background()

	if err := w.pvc(ctx); err != nil {
		return err
	}

	if err := w.deployment(ctx); err != nil {
		return err
	}

	if err := w.service(ctx); err != nil {
		return err
	}
	return nil
}

func (w *wordpress) service(ctx context.Context) error {
	svcCli := w.cliSet.CoreV1().Services(w.namespace)

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      w.name,
			Namespace: w.namespace,
			Labels: map[string]string{
				"app":                    w.name,
				"app.kubernetes.io/name": w.name,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":  w.name,
				"tier": "frontend",
			},
			Ports: []corev1.ServicePort{
				{
					Name:     w.name,
					Protocol: corev1.ProtocolTCP,
					Port:     w.port,
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

func (w *wordpress) pvc(ctx context.Context) error {
	pvcCli := w.cliSet.CoreV1().PersistentVolumeClaims(w.namespace)

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      w.name,
			Namespace: w.namespace,
			Labels: map[string]string{
				"app": w.name,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(w.wwwVolumeSize),
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

func (w *wordpress) deployment(ctx context.Context) error {
	dCli := w.cliSet.AppsV1().Deployments(w.namespace)

	deployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      w.name,
			Namespace: w.namespace,
			Labels: map[string]string{
				"app": w.name,
			},
		},
		Spec: v1.DeploymentSpec{
			Replicas: &w.replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  w.name,
					"tier": "frontend",
				},
			},
			Strategy: v1.DeploymentStrategy{
				Type: v1.RecreateDeploymentStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  w.name,
						"tier": "frontend",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  w.name,
							Image: w.image,
							Env: []corev1.EnvVar{
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
							Resources: w.resources,
							Ports: []corev1.ContainerPort{
								{
									Name:          w.name,
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: w.port,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      w.name,
									MountPath: "/var/www/html",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: w.name,
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: w.name,
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
