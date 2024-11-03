package app

import (
	"context"

	"golang.org/x/exp/rand"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (a *Application) secret(ctx context.Context) error {
	scrtCli := a.Client.CoreV1().Secrets(a.Config.Namespace)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.Config.Secret.SecretName,
			Namespace: a.Config.Namespace,
			Labels:    a.Config.Labels,
		},
		Type: a.Config.Secret.SecretType,
		Data: a.Config.Secret.SecretData,
	}

	_, err := scrtCli.Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			if _, err := scrtCli.Update(ctx, secret, metav1.UpdateOptions{}); err != nil {
				return err
			}
			return nil
		}
		return err
	}
	return nil
}

func GeneratePassword(length int) []byte {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	pass := make([]byte, length)
	for i := range pass {
		pass[i] = charset[rand.Intn(len(charset))]
	}
	return pass
}
