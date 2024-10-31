/*
Copyright © 2024 Kol Ratner kolratner@gmail.com
*/
package cmd

import (
	"log"

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
	Run: deployEntrypoint,
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

func deployEntrypoint(cmd *cobra.Command, args []string) {
	msgs := make(chan string)
	// the done channel signals to the main goroutine that the apps.Deploy() function has completed
	// otherwise our program will continue trying to process messages from the apps.Deploy() function and panic
	done := make(chan bool)

	go func() {
		if err := apps.Deploy(msgs); err != nil {
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
