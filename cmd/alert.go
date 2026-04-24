package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

var alertCmd = &cobra.Command{
	Use:   "alert",
	Short: "Manage alert rules",
	Long:  "Adds, lists, enables, disables, removes, and inspects alert history records.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(M4-4): implement alert command root behavior.
		return errors.New("not implemented")
	},
}

var alertAddCmd = &cobra.Command{
	Use:   "add <TICKER>",
	Short: "Add an alert rule",
	Long:  "Parses an alert condition expression and creates a new alert rule for a ticker.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(M4-4): implement alert add command.
		return errors.New("not implemented")
	},
}

var alertListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured alert rules",
	Long:  "Displays alert rules with condition, channel, active status, and cooldown information.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(M4-4): implement alert list command.
		return errors.New("not implemented")
	},
}

var alertRemoveCmd = &cobra.Command{
	Use:   "remove <ID>",
	Short: "Remove an alert rule",
	Long:  "Deletes an alert rule by its identifier.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(M4-4): implement alert remove command.
		return errors.New("not implemented")
	},
}

var alertEnableCmd = &cobra.Command{
	Use:   "enable <ID>",
	Short: "Enable an alert rule",
	Long:  "Sets an alert rule active without modifying any other rule attributes.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(M4-4): implement alert enable command.
		return errors.New("not implemented")
	},
}

var alertDisableCmd = &cobra.Command{
	Use:   "disable <ID>",
	Short: "Disable an alert rule",
	Long:  "Sets an alert rule inactive without deleting it.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(M4-4): implement alert disable command.
		return errors.New("not implemented")
	},
}

var alertHistoryCmd = &cobra.Command{
	Use:   "history",
	Short: "Show historical alert events",
	Long:  "Displays alert trigger history records in reverse chronological order.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(M4-4): implement alert history command.
		return errors.New("not implemented")
	},
}

func init() {
	alertCmd.AddCommand(alertAddCmd)
	alertCmd.AddCommand(alertListCmd)
	alertCmd.AddCommand(alertRemoveCmd)
	alertCmd.AddCommand(alertEnableCmd)
	alertCmd.AddCommand(alertDisableCmd)
	alertCmd.AddCommand(alertHistoryCmd)
	rootCmd.AddCommand(alertCmd)
}
