package apps

import (
	"github.com/kol-ratner/tufin/internal/apps/mysql"
	"github.com/kol-ratner/tufin/internal/apps/wordpress"
	"github.com/kol-ratner/tufin/internal/k8s"
)

func Deploy(msgChan chan<- string) error {
	// grabbing the kubeconfig from the hosts default ~/.kube/config file
	kubeConfig, err := k8s.GetKubeConfigFromHost("")
	if err != nil {
		return err
	}
	msgChan <- "found kubeconfig!"

	clientSet, err := k8s.NewClient(kubeConfig)
	if err != nil {
		return err
	}

	mysql := mysql.New(clientSet, &mysql.Options{
		CPURequest:    "500m",
		MemoryRequest: "512Mi",
	})
	if err := mysql.Deploy(); err != nil {
		return err
	}
	msgChan <- "succesfully triggered mysql deployment"

	wp := wordpress.New(clientSet, nil)
	if err := wp.Deploy(); err != nil {
		return err
	}
	msgChan <- "succesfully triggered wordpress deployment"

	return nil
}
