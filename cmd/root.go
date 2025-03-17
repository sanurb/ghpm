package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// rootCmd is the main Cobra command.
var rootCmd = &cobra.Command{
	Use:   "ghpm",
	Short: "ghpm - GitHub Project Manager",
	// On no subcommand, launch the interactive TUI.
	Run: func(cmd *cobra.Command, args []string) {
		if err := InteractiveCmd(); err != nil {
			fmt.Println("Error:", err)
		}
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
