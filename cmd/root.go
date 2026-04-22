package cmd

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tkr",
	Short: "Terminal stock monitor and alert system",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	rootCmd.PersistentFlags().String("config", "", "Config file path (default: ~/.config/tkr/config.yaml)")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level: debug, info, warn, error")
}
