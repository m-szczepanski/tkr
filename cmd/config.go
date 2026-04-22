package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage application configuration",
	Long:  "Shows, updates, and validates the resolved tkr configuration.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(M8-1): implement config command root behavior.
		return errors.New("not implemented")
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show resolved configuration",
	Long:  "Prints the active configuration with sensitive values masked.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(M8-1): implement config show command.
		return errors.New("not implemented")
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <KEY> <VALUE>",
	Short: "Set a single configuration value",
	Long:  "Updates one key in the user configuration file.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(M8-1): implement config set command.
		return errors.New("not implemented")
	},
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configured providers",
	Long:  "Checks enabled provider configuration and reports connectivity and auth status.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(M8-2): implement config validate command.
		return errors.New("not implemented")
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configValidateCmd)
	rootCmd.AddCommand(configCmd)
}
