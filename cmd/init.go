package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yourname/tkr/internal/apperrors"
	"github.com/yourname/tkr/internal/config"
	"github.com/yourname/tkr/internal/db"
)

const defaultInitConfigPath = "~/.config/tkr/config.yaml"

var initCmd = &cobra.Command{
	Use:     "init",
	Short:   "Initialise local configuration and database",
	Long:    "Creates the config file, initialises the SQLite database, and applies pending migrations.",
	Example: "tkr init",
	RunE: func(cmd *cobra.Command, args []string) error {
		flagConfigPath, err := cmd.Flags().GetString("config")
		if err != nil {
			return fmt.Errorf("read --config flag: %w", errors.Join(apperrors.ErrConfig, err))
		}

		configPath, err := resolveInitConfigPath(flagConfigPath)
		if err != nil {
			return fmt.Errorf("resolve config path: %w", errors.Join(apperrors.ErrConfig, err))
		}

		createdConfig, err := ensureUserConfig(configPath)
		if err != nil {
			return fmt.Errorf("prepare config file: %w", errors.Join(apperrors.ErrConfig, err))
		}

		prevConfigEnv, hadConfigEnv := os.LookupEnv("TKR_CONFIG")
		if err := os.Setenv("TKR_CONFIG", configPath); err != nil {
			return fmt.Errorf("set TKR_CONFIG: %w", errors.Join(apperrors.ErrConfig, err))
		}
		defer func() {
			if hadConfigEnv {
				_ = os.Setenv("TKR_CONFIG", prevConfigEnv)
				return
			}
			_ = os.Unsetenv("TKR_CONFIG")
		}()

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("load configuration: %w", err)
		}

		repo, err := db.Open(cfg.Database.Path)
		if err != nil {
			return fmt.Errorf("initialise database: %w", errors.Join(apperrors.ErrDB, err))
		}

		defer func() {
			_ = repo.Close()
		}()

		cmd.Println("tkr initialization complete.")
		if createdConfig {
			cmd.Printf("Created config: %s\n", configPath)
		} else {
			cmd.Printf("Config already exists: %s\n", configPath)
		}
		cmd.Printf("Database ready: %s\n", cfg.Database.Path)
		cmd.Println("\nNext steps:")
		cmd.Println("  tkr watch add AAPL")
		cmd.Println("  tkr quote AAPL")
		cmd.Println("  tkr daemon start --foreground")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func resolveInitConfigPath(flagPath string) (string, error) {
	path := flagPath
	if strings.TrimSpace(path) == "" {
		path = os.Getenv("TKR_CONFIG")
	}
	if strings.TrimSpace(path) == "" {
		path = defaultInitConfigPath
	}

	return expandTilde(path)
}

func ensureUserConfig(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, err
	}

	template, err := readConfigTemplate()
	if err != nil {
		return false, err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return false, nil
		}
		return false, err
	}
	defer func() {
		_ = file.Close()
	}()

	if _, err := file.Write(template); err != nil {
		return false, err
	}

	return true, nil
}

func readConfigTemplate() ([]byte, error) {
	const templateName = "config.example.yaml"

	if content, err := os.ReadFile(templateName); err == nil {
		return content, nil
	}

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("resolve init command path")
	}

	templatePath := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", templateName))
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", templateName, err)
	}

	return content, nil
}

func expandTilde(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if path == "~" {
		return homeDir, nil
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir, path[2:]), nil
	}

	return path, nil
}
