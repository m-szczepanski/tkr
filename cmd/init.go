package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialise local configuration and database",
	Long:  "Creates the config file, initialises the SQLite database, and applies pending migrations.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(M1-7): implement init command workflow.
		return errors.New("not implemented")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
