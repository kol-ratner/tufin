package deployments

import (
	"github.com/kol-ratner/tufin/internal/deployments/mysql"
	"github.com/kol-ratner/tufin/internal/deployments/wordpress"

	// "github.com/kol-ratner/tufin/internal/deployments/wordpress"
	"github.com/kol-ratner/tufin/internal/k8s"
)

func Ship(msgChan chan<- string) error {
	// grabbing the kubeconfig from the hosts default ~/.kube/config file
	kubeConfig, err := k8s.GetKubeConfigFromHost("")
	if err != nil {
		return err
	}
	msgChan <- "found kubeconfig!"

	cli, err := k8s.NewClient(kubeConfig)
	if err != nil {
		return err
	}

	mysql := mysql.New(
		cli.ClientSet,
		mysql.WithCPURequest("250m"),
		mysql.WithMemoryRequest("250Mi"),
	)
	if err := mysql.Deploy(); err != nil {
		return err
	}

	msgChan <- "succesfully triggered mysql deployment"

	wp := wordpress.New(cli.ClientSet)
	if err := wp.Deploy(); err != nil {
		return err
	}
	msgChan <- "succesfully triggered wordpress deployment"

	return nil
}
