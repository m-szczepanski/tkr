package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Manage the polling daemon",
	Long:  "Starts, stops, checks status, and restarts the background polling daemon.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(M7-3): implement daemon command root behavior.
		return errors.New("not implemented")
	},
}

var daemonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the daemon",
	Long:  "Starts the scheduler loop in foreground or as a background process.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(M7-3): implement daemon start command.
		return errors.New("not implemented")
	},
}

var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the daemon",
	Long:  "Stops the running daemon process using its PID file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(M7-3): implement daemon stop command.
		return errors.New("not implemented")
	},
}

var daemonStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show daemon status",
	Long:  "Displays daemon process state, uptime, and latest polling statistics.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(M7-3): implement daemon status command.
		return errors.New("not implemented")
	},
}

var daemonRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the daemon",
	Long:  "Stops and starts the daemon sequentially.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(M7-3): implement daemon restart command.
		return errors.New("not implemented")
	},
}

func init() {
	daemonCmd.AddCommand(daemonStartCmd)
	daemonCmd.AddCommand(daemonStopCmd)
	daemonCmd.AddCommand(daemonStatusCmd)
	daemonCmd.AddCommand(daemonRestartCmd)
	rootCmd.AddCommand(daemonCmd)
}
