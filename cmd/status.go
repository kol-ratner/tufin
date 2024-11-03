/*
Copyright © 2024 Kol Ratner kolratner@gmail.com
*/
package cmd

import (
	"log"

	"github.com/kol-ratner/tufin/internal/reporting"
	"github.com/kol-ratner/tufin/pkg/k8s"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check deployment status of WordPress and MySQL applications",
	Long: `The status command provides real-time information about your deployed applications.

Displays:
  - Pod status and health
  - Resource utilization

Examples:
  # Get status of all deployments
  tufin status

  # View detailed resource usage
  tufin status`,
	Run: statusEntrypoint,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func statusEntrypoint(cmd *cobra.Command, args []string) {
	msgs := make(chan string)
	// the done channel signals to the main goroutine that the cluster.Create() function has completed
	// otherwise our program will continue trying to process messages from the cluster.Create() function and panic
	done := make(chan bool)

	go func() {
		// grabbing the kubeconfig from the hosts default ~/.kube/config file
		kubeConfig, err := k8s.GetKubeConfigFromHost("")
		if err != nil {
			log.Fatal(err)
		}

		cli, err := k8s.NewClient(kubeConfig)
		if err != nil {
			log.Fatal(err)
		}

		if err := reporting.Status(msgs, "", cli); err != nil {
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
