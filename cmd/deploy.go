/*
Copyright © 2024 Kol Ratner kolratner@gmail.com
*/
package cmd

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kol-ratner/tufin/internal/config"
	"github.com/kol-ratner/tufin/internal/deployments"
	"github.com/kol-ratner/tufin/pkg/k8s"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy applications to kubernetes",
	Long: `Deploy WordPress and MySQL applications to Kubernetes with configurable resources.

The deploy command supports customizing resource allocations for both WordPress and MySQL components
using dot notation to specify which component gets which settings.

Available Components:
  - wordpress
  - mysql

Configuration Options:
  - replicas      : Number of pod replicas (int)
  - cpu-request   : Minimum CPU guaranteed (e.g. 250m, 500m)
  - memory-request: Minimum memory guaranteed (e.g. 256Mi, 1Gi)
  - cpu-limit     : Maximum CPU allowed (e.g. 500m, 1)
  - memory-limit  : Maximum memory allowed (e.g. 512Mi, 2Gi)
  - volume-size   : Persistent volume size (e.g. 5Gi, 10Gi)

Examples:
  # Deploy WordPress with 2 replicas and MySQL with 3 replicas
  tufin deploy --set wordpress.replicas=2,mysql.replicas=3

  # Deploy WordPress with custom memory and volume size
  tufin deploy --set wordpress.memory-request=1Gi,wordpress.volume-size=10Gi

  # Full deployment with multiple configurations
  tufin deploy --set wordpress.replicas=2,wordpress.memory-request=1Gi,mysql.replicas=3,mysql.cpu-request=500m`,
	Run: deployEntrypoint,
}

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().String("set", "", `Available options (comma-separated key:value pairs):
  replicas        - Number of replicas (int)
  cpu-request     - CPU request (e.g. 250m, 500m)
  memory-request  - Memory request (e.g. 256Mi, 1Gi)
  cpu-limit       - CPU limit (e.g. 500m, 1)
  memory-limit    - Memory limit (e.g. 512Mi, 2Gi)
  volume-size     - Volume size (e.g. 5Gi, 10Gi)

Example: --set wordpress.replicas=2,wordpress.volume-size=1Gi,mysql.replicas=3
`)
}

func deployEntrypoint(cmd *cobra.Command, args []string) {
	setValue, err := cmd.Flags().GetString("set")
	if err != nil {
		log.Fatal(err)
	}

	componentOpts, err := ParseSetFlag(setValue)
	if err != nil {
		log.Fatal(err)
	}

	// Convert to configs slice
	var configs []deployments.DeploymentConfig
	for component, opts := range componentOpts {
		configs = append(configs, deployments.DeploymentConfig{
			Component: component,
			Options:   opts,
		})
	}

	msgs := make(chan string)
	kubeConfig, err := k8s.GetKubeConfigFromHost("")
	if err != nil {
		log.Fatal(err)
	}
	msgs <- "found kubeconfig!"

	cli, err := k8s.NewClient(kubeConfig)
	if err != nil {
		log.Fatal(err)
	}

	// the done channel signals to the main goroutine that the apps.Deploy() function has completed
	// otherwise our program will continue trying to process messages from the apps.Deploy() function and panic
	done := make(chan bool)

	go func() {
		if err := deployments.Ship(msgs, "", cli, configs...); err != nil {
			log.Println(err)
		}
		done <- true
	}()

	for {
		select {
		case msg := <-msgs:
			log.Println(msg)
		case <-done:
			close(msgs)
			return
		}
	}
}

func ParseSetFlag(setValue string) (map[string][]config.Option, error) {
	componentOpts := make(map[string][]config.Option)
	pairs := strings.Split(setValue, ",")

	for _, pair := range pairs {
		parts := strings.Split(pair, ".")
		if len(parts) != 2 {
			continue
		}

		component := parts[0]
		// Parse the key=value part
		kvPair := strings.Split(parts[1], "=")
		if len(kvPair) != 2 {
			continue
		}

		key, value := kvPair[0], kvPair[1]
		opt, err := parseOption(key, value)
		if err != nil {
			return nil, err
		}
		componentOpts[component] = append(componentOpts[component], opt)
	}

	return componentOpts, nil
}

func parseOption(key, value string) (config.Option, error) {
	switch key {
	case "replicas":
		replicaInt, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("invalid value type for replicas: %s", value)
		}
		replicas := int32(replicaInt)
		return config.WithReplicas(replicas), nil
	case "cpu-request":
		return config.WithCPURequest(value), nil
	case "memory-request":
		return config.WithMemoryRequest(value), nil
	case "cpu-limit":
		return config.WithCPULimit(value), nil
	case "memory-limit":
		return config.WithMemoryLimit(value), nil
	case "volume-size":
		return config.WithVolumeSize(value), nil
	default:
		return nil, fmt.Errorf("invalid option: %s", key)
	}
}
