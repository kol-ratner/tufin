package mysql

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
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
		Replicas:      3,
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
		image:     "mysql",
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

	if err := m.configMap(ctx); err != nil {
		return err
	}

	if err := m.statefulSet(ctx); err != nil {
		return err
	}

	if err := m.service(ctx); err != nil {
		return err
	}
	return nil
}

func (m *mysql) configMap(ctx context.Context) error {
	cmCli := m.cliSet.CoreV1().ConfigMaps(m.namespace)

	primaryConfig := "[mysqld]\nlog-bin"
	replicaConfig := "[mysqld]\nsuper-read-only"
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.name,
			Namespace: m.namespace,
		},
		Data: map[string]string{
			"primary.cnf": primaryConfig,
			"replica.cnf": replicaConfig,
		},
	}

	_, err := cmCli.Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			cmCli.Update(ctx, cm, metav1.UpdateOptions{})
			return nil
		}
		return err
	}
	return nil
}

func (m *mysql) service(ctx context.Context) error {
	svcCli := m.cliSet.CoreV1().Services(m.namespace)

	headless := &corev1.Service{
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
				"app": m.name,
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

	readOnly := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-read", m.name),
			Namespace: m.namespace,
			Labels: map[string]string{
				"app":                    m.name,
				"app.kubernetes.io/name": m.name,
				"readonly":               "true",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": m.name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:     m.name,
					Protocol: corev1.ProtocolTCP,
					Port:     m.port,
				},
			},
		},
	}

	services := []*corev1.Service{
		headless,
		readOnly,
	}

	for _, svc := range services {
		_, err := svcCli.Create(ctx, svc, metav1.CreateOptions{})
		if err != nil {
			if apierrors.IsAlreadyExists(err) {
				svcCli.Update(ctx, svc, metav1.UpdateOptions{})
				return nil
			}
			return err
		}
	}

	return nil
}

func (m *mysql) statefulSet(ctx context.Context) error {
	ssCli := m.cliSet.AppsV1().StatefulSets(m.namespace)

	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.name,
			Namespace: m.namespace,
		},

		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":                    m.name,
					"app.kubernetes.io/name": m.name,
				},
			},
			ServiceName: m.name,
			Replicas:    &m.replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                    m.name,
						"app.kubernetes.io/name": m.name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  m.name,
							Image: m.image,
							Env: []corev1.EnvVar{
								{
									Name:  "MYSQL_ALLOW_EMPTY_PASSWORD",
									Value: "1",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          m.name,
									ContainerPort: m.port,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Resources: m.resources,
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"mysqladmin",
											"ping",
										},
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       2,
								TimeoutSeconds:      1,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"mysqladmin",
											"-h",
											"127.0.0.1",
											"-e",
											"SELECT 1",
										},
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       2,
								TimeoutSeconds:      1,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/var/lib/mysql",
									SubPath:   m.name,
								},
								{
									Name:      "conf",
									MountPath: "/etc/mysql/conf.d",
									SubPath:   m.name,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "conf",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "config-map",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: m.name,
									},
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("10Gi"),
							},
						},
					},
				},
			},
		},
	}

	_, err := ssCli.Create(ctx, statefulSet, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			ssCli.Update(ctx, statefulSet, metav1.UpdateOptions{})
			return nil
		}
		return err

	}

	return nil
}
