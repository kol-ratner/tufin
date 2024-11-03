package app

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (a *Application) pvc(ctx context.Context) error {
	pvcCli := a.Client.CoreV1().PersistentVolumeClaims(a.Config.Namespace)

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.Config.Name,
			Namespace: a.Config.Namespace,
			Labels:    a.Config.Labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				a.Config.Pvc.AccessMode,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(a.Config.Pvc.Size),
				},
			},
		},
	}

	_, err := pvcCli.Create(ctx, pvc, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			if _, err := pvcCli.Update(ctx, pvc, metav1.UpdateOptions{}); err != nil {
				return err
			}
			return nil
		}
		return err
	}

	return nil
}
