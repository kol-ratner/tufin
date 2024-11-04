package deployments

import (
	"fmt"

	"k8s.io/client-go/kubernetes"

	"github.com/kol-ratner/tufin/internal/config"
	"github.com/kol-ratner/tufin/internal/deployments/mysql"
	"github.com/kol-ratner/tufin/internal/deployments/wordpress"
)

type DeploymentConfig struct {
	Component string
	Options   []config.Option
}

func Ship(msgChan chan<- string, cli kubernetes.Interface, configs ...DeploymentConfig) error {

	// If no configs provided, deploy everything with defaults
	if len(configs) == 0 {
		return deployAll(cli, msgChan)
	}

	// Deploy selected components with their options
	for _, cfg := range configs {
		switch cfg.Component {
		case "mysql":
			mysql := mysql.New(cli, cfg.Options...)
			if err := mysql.Deploy(); err != nil {
				return err
			}
			msgChan <- "successfully triggered mysql deployment"

		case "wordpress":
			wp := wordpress.New(cli, cfg.Options...)
			if err := wp.Deploy(); err != nil {
				return err
			}
			msgChan <- "successfully triggered wordpress deployment"
		default:
			return fmt.Errorf("unsupported component: %s", cfg.Component)
		}
	}
	return nil
}

func deployAll(cli kubernetes.Interface, msgChan chan<- string) error {
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
