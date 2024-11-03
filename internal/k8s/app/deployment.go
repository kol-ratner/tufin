package app

import (
	"context"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (a *Application) deployment(ctx context.Context) error {
	dCli := a.Client.AppsV1().Deployments(a.Config.Namespace)

	deployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.Config.Name,
			Namespace: a.Config.Namespace,
			Labels:    a.Config.Labels,
		},
		Spec: v1.DeploymentSpec{
			Replicas: &a.Config.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": a.Config.Name,
				},
			},
			Strategy: v1.DeploymentStrategy{
				Type: v1.RecreateDeploymentStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: a.Config.Labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:      a.Config.Name,
							Image:     a.Config.Image,
							Env:       a.Config.EnvVars,
							Resources: a.Config.Resources,
							Ports: []corev1.ContainerPort{
								{
									Name:          a.Config.Name,
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: a.Config.ContainerPort,
								},
							},
							VolumeMounts: a.Config.VolumeMounts,
						},
					},
					Volumes: a.Config.Volumes,
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
