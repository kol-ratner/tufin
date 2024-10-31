/*
Copyright © 2024 Kol Ratner kolratner@gmail.com
*/
package cmd

import (
	"log"

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
	Run: clusterEntrypoint,
}

func init() {
	rootCmd.AddCommand(clusterCmd)

	// clusterCmd.Flags().IntP("nodes", "n", 0, "Specify how many nodes you want to create")
	// clusterCmd.Flags().String("name", "default", "Specify a name for the cluster you want to create")
}

func clusterEntrypoint(cmd *cobra.Command, args []string) {
	msgs := make(chan string)
	// the done channel signals to the main goroutine that the cluster.Create() function has completed
	// otherwise our program will continue trying to process messages from the cluster.Create() function and panic
	done := make(chan bool)

	go func() {
		if err := cluster.Create(msgs); err != nil {
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
