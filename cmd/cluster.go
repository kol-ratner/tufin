/*
Copyright © 2024 Kol Ratner kolratner@gmail.com
*/
package cmd

import (
	"github.com/kol-ratner/tufin/internal/cluster"
	"github.com/spf13/cobra"
)

// clusterCmd represents the cluster command
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cluster.Create(); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(clusterCmd)

	clusterCmd.Flags().IntP("nodes", "n", 0, "Specify how many nodes you want to create")

	clusterCmd.Flags().String("name", "default", "Specify a name for the cluster you want to create")
}
