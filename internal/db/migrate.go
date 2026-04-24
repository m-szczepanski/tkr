package db

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"

	"github.com/jmoiron/sqlx"
)

const schemaMigrationsDDL = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	filename   TEXT PRIMARY KEY,
	applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);`

// Migrate applies pending SQL migrations.
func Migrate(db *sqlx.DB) error {
	if db == nil {
		return fmt.Errorf("db.Migrate: nil database")
	}

	if _, err := db.Exec(schemaMigrationsDDL); err != nil {
		return fmt.Errorf("db.Migrate: ensure schema_migrations table: %w", err)
	}

	dir, err := findMigrationsDir()
	if err != nil {
		return fmt.Errorf("db.Migrate: find migrations directory: %w", err)
	}

	files, err := listMigrationFiles(dir)
	if err != nil {
		return fmt.Errorf("db.Migrate: list migration files: %w", err)
	}

	for _, file := range files {
		filename := filepath.Base(file)

		applied, err := isMigrationApplied(db, filename)
		if err != nil {
			return fmt.Errorf("db.Migrate: check migration %q: %w", filename, err)
		}
		if applied {
			continue
		}

		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("db.Migrate: read migration %q: %w", filename, err)
		}

		if _, err := db.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("db.Migrate: apply migration %q: %w", filename, err)
		}

		if _, err := db.Exec(`INSERT INTO schema_migrations (filename) VALUES (?)`, filename); err != nil {
			return fmt.Errorf("db.Migrate: record migration %q: %w", filename, err)
		}
	}

	return nil
}

func findMigrationsDir() (string, error) {
	const migrationsDirName = "migrations"

	if info, err := os.Stat(migrationsDirName); err == nil && info.IsDir() {
		return migrationsDirName, nil
	}

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("resolve caller path")
	}

	projectRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	fallbackDir := filepath.Join(projectRoot, migrationsDirName)

	if info, err := os.Stat(fallbackDir); err == nil && info.IsDir() {
		return fallbackDir, nil
	}

	return "", fmt.Errorf("migrations directory not found")
}

func listMigrationFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}

		files = append(files, filepath.Join(dir, entry.Name()))
	}

	sort.Strings(files)
	return files, nil
}

func isMigrationApplied(db *sqlx.DB, filename string) (bool, error) {
	var count int
	if err := db.Get(&count, `SELECT COUNT(1) FROM schema_migrations WHERE filename = ?`, filename); err != nil {
		return false, err
	}

	return count > 0, nil
}
