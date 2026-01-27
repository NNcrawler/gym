package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "gym",
	Short: "gym manages synchronization of agent skills into projects",
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(addCmd())
	rootCmd.AddCommand(removeCmd())
	rootCmd.AddCommand(syncCmd())
}
