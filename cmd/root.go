package cmd

import "github.com/spf13/cobra"

// rootCmd is the main Cobra command.
var rootCmd = &cobra.Command{
	Use:   "ghpm",
	Short: "ghpm - GitHub Project Manager",
	// When no subcommand is provided, launch the interactive UI.
	Run: func(cmd *cobra.Command, args []string) {
		InteractiveCmd()
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
