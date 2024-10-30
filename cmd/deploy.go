/*
Copyright © 2024 Kol Ratner kolratner@gmail.com
*/
package cmd

import (
	"github.com/kol-ratner/tufin/internal/apps"
	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := apps.Deploy(); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
