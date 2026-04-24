package cmd

import (
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tkr",
	Short: "Terminal stock monitor and alert system",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	rootCmd.PersistentFlags().String("config", "", "Config file path (default: ~/.config/tkr/config.yaml)")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level: debug, info, warn, error")
}
