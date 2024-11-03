package deployments

import (
	"k8s.io/client-go/kubernetes"

	"github.com/kol-ratner/tufin/internal/config"
	"github.com/kol-ratner/tufin/internal/deployments/mysql"
	"github.com/kol-ratner/tufin/internal/deployments/wordpress"

	"github.com/kol-ratner/tufin/pkg/k8s"
)

type DeploymentConfig struct {
	Component string
	Options   []config.Option
}

func Ship(msgChan chan<- string, configs ...DeploymentConfig) error {
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

	// If no configs provided, deploy everything with defaults
	if len(configs) == 0 {
		return deployAll(cli.ClientSet, msgChan)
	}

	// Deploy selected components with their options
	for _, cfg := range configs {
		switch cfg.Component {
		case "mysql":
			mysql := mysql.New(cli.ClientSet, cfg.Options...)
			if err := mysql.Deploy(); err != nil {
				return err
			}
			msgChan <- "successfully triggered mysql deployment"

		case "wordpress":
			wp := wordpress.New(cli.ClientSet, cfg.Options...)
			if err := wp.Deploy(); err != nil {
				return err
			}
			msgChan <- "successfully triggered wordpress deployment"
		}
	}
	return nil
}

func deployAll(cli *kubernetes.Clientset, msgChan chan<- string) error {
	mysql := mysql.New(cli)
	if err := mysql.Deploy(); err != nil {
		return err
	}
	msgChan <- "successfully triggered mysql deployment"

	wp := wordpress.New(cli)
	if err := wp.Deploy(); err != nil {
		return err
	}
	msgChan <- "successfully triggered wordpress deployment"

	return nil
}
