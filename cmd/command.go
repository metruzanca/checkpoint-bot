package cmd

import (
	"github.com/spf13/cobra"
)

// Root command for managing commands
var commandCmd = &cobra.Command{
	Use:   "command",
	Short: "Manage commands",
}

func init() {
	rootCmd.AddCommand(commandCmd)
}
