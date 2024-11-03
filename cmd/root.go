/*
Copyright © 2024 Kol Ratner kolratner@gmail.com
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tufin",
	Short: "Kubernetes deployment tool for WordPress and MySQL applications",
	Long: `Tufin is a powerful CLI tool for deploying and managing WordPress and MySQL on Kubernetes.

Key Features:
  - One-command deployment of WordPress and MySQL
  - Customizable resource allocation
  - Real-time deployment status monitoring

Core Commands:
  deploy    Deploy applications with custom configurations
  status    Monitor deployment health and status
  cluster   Manage Kubernetes cluster settings

Getting started:
  tufin cluster
  tufin deploy --set wordpress.replicas=2
  tufin status`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
