package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

var quoteCmd = &cobra.Command{
	Use:   "quote <TICKER> [TICKER...]",
	Short: "Fetch live quotes for one or more tickers",
	Long:  "Retrieves live quote data and optionally renders recent history as terminal sparklines.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(M2-10): implement quote command.
		return errors.New("not implemented")
	},
}

func init() {
	rootCmd.AddCommand(quoteCmd)
}
