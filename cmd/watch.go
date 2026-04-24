package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Manage the stock watchlist",
	Long:  "Adds, removes, and lists watched stocks that are polled by the daemon.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(M3-2): implement watch command root behavior.
		return errors.New("not implemented")
	},
}

var watchAddCmd = &cobra.Command{
	Use:   "add <TICKER>",
	Short: "Add a stock to the watchlist",
	Long:  "Validates a ticker with providers and stores it in the watchlist table.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(M3-2): implement watch add command.
		return errors.New("not implemented")
	},
}

var watchRemoveCmd = &cobra.Command{
	Use:   "remove <TICKER>",
	Short: "Remove a stock from the watchlist",
	Long:  "Removes a watched stock and optionally purges associated alert rules.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(M3-2): implement watch remove command.
		return errors.New("not implemented")
	},
}

var watchListCmd = &cobra.Command{
	Use:   "list",
	Short: "List watched stocks with current quotes",
	Long:  "Displays watchlist entries and current quote data in table, JSON, or CSV formats.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(M3-2): implement watch list command.
		return errors.New("not implemented")
	},
}

func init() {
	watchCmd.AddCommand(watchAddCmd)
	watchCmd.AddCommand(watchRemoveCmd)
	watchCmd.AddCommand(watchListCmd)
	rootCmd.AddCommand(watchCmd)
}
