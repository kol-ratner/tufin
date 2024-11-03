package app

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (a *Application) service(ctx context.Context) error {
	svcCli := a.Client.CoreV1().Services(a.Config.Namespace)

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.Config.Name,
			Namespace: a.Config.Namespace,
			Labels:    a.Config.Labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: a.Config.Deployment.SelectorMatchLabels,
			Ports: []corev1.ServicePort{
				{
					Name:     a.Config.Name,
					Protocol: corev1.ProtocolTCP,
					Port:     a.Config.Svc.Port,
				},
			},
		},
	}

	// Set ClusterIP to "None" for headless service
	if a.Config.Svc.DisableClusterIP {
		svc.Spec.ClusterIP = "None"
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
