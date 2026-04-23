package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"

	"github.com/yourname/tkr/pkg/models"
)

var ErrNotImplemented = errors.New("not implemented")

// SQLiteRepository is the SQLite-backed Repository implementation.
type SQLiteRepository struct {
	db *sqlx.DB
}

// Open opens a SQLite repository at the given path.
func Open(path string) (Repository, error) {
	expandedPath, err := expandDBPath(path)
	if err != nil {
		return nil, fmt.Errorf("db.Open: expand path: %w", err)
	}

	if err := ensureParentDir(expandedPath); err != nil {
		return nil, fmt.Errorf("db.Open: ensure db directory: %w", err)
	}

	sqlDB, err := sqlx.Open("sqlite", expandedPath)
	if err != nil {
		return nil, fmt.Errorf("db.Open: open sqlite: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("db.Open: ping sqlite: %w", err)
	}

	if _, err := sqlDB.Exec(`
		PRAGMA journal_mode=WAL;
		PRAGMA foreign_keys=ON;
		PRAGMA busy_timeout=5000;
	`); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("db.Open: set pragmas: %w", err)
	}

	if err := Migrate(sqlDB); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("db.Open: run migrations: %w", err)
	}

	return &SQLiteRepository{db: sqlDB}, nil
}

func (r *SQLiteRepository) AddStock(ctx context.Context, s models.Stock) error {
	_, _ = ctx, s
	return fmt.Errorf("db.AddStock: %w", ErrNotImplemented)
}

func (r *SQLiteRepository) RemoveStock(ctx context.Context, ticker string) error {
	_, _ = ctx, ticker
	return fmt.Errorf("db.RemoveStock: %w", ErrNotImplemented)
}

func (r *SQLiteRepository) ListStocks(ctx context.Context) ([]models.Stock, error) {
	_ = ctx
	return nil, fmt.Errorf("db.ListStocks: %w", ErrNotImplemented)
}

func (r *SQLiteRepository) GetStock(ctx context.Context, ticker string) (models.Stock, error) {
	_, _ = ctx, ticker
	return models.Stock{}, fmt.Errorf("db.GetStock: %w", ErrNotImplemented)
}

func (r *SQLiteRepository) AddAlertRule(ctx context.Context, rule models.AlertRule) (int64, error) {
	_, _ = ctx, rule
	return 0, fmt.Errorf("db.AddAlertRule: %w", ErrNotImplemented)
}

func (r *SQLiteRepository) RemoveAlertRule(ctx context.Context, id int64) error {
	_, _ = ctx, id
	return fmt.Errorf("db.RemoveAlertRule: %w", ErrNotImplemented)
}

func (r *SQLiteRepository) ListAlertRules(ctx context.Context, ticker string) ([]models.AlertRule, error) {
	_, _ = ctx, ticker
	return nil, fmt.Errorf("db.ListAlertRules: %w", ErrNotImplemented)
}

func (r *SQLiteRepository) UpdateAlertRule(ctx context.Context, rule models.AlertRule) error {
	_, _ = ctx, rule
	return fmt.Errorf("db.UpdateAlertRule: %w", ErrNotImplemented)
}

func (r *SQLiteRepository) AddAlertEvent(ctx context.Context, event models.AlertEvent) error {
	_, _ = ctx, event
	return fmt.Errorf("db.AddAlertEvent: %w", ErrNotImplemented)
}

func (r *SQLiteRepository) ListAlertEvents(ctx context.Context, filter models.AlertEventFilter) ([]models.AlertEvent, error) {
	_, _ = ctx, filter
	return nil, fmt.Errorf("db.ListAlertEvents: %w", ErrNotImplemented)
}

func (r *SQLiteRepository) SaveQuote(ctx context.Context, q models.Quote) error {
	recordedAt := q.Timestamp.UTC()
	if q.Timestamp.IsZero() {
		recordedAt = time.Now().UTC()
	}

	const query = `
		INSERT INTO quote_history (ticker, price, volume, change_pct, source, recorded_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	if _, err := r.db.ExecContext(ctx, query, q.Ticker, q.Price, q.Volume, q.ChangePct, q.Source, recordedAt); err != nil {
		return fmt.Errorf("db.SaveQuote: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) GetRecentQuotes(ctx context.Context, ticker string, limit int) ([]models.Quote, error) {
	if limit <= 0 {
		return []models.Quote{}, nil
	}

	const query = `
		SELECT ticker, price, volume, change_pct, source, recorded_at AS timestamp
		FROM quote_history
		WHERE ticker = ?
		ORDER BY recorded_at DESC
		LIMIT ?
	`

	var quotes []models.Quote
	if err := r.db.SelectContext(ctx, &quotes, query, ticker, limit); err != nil {
		return nil, fmt.Errorf("db.GetRecentQuotes: %w", err)
	}

	return quotes, nil
}

func (r *SQLiteRepository) PruneHistory(ctx context.Context, olderThan time.Duration) error {
	if olderThan <= 0 {
		return nil
	}

	cutoff := time.Now().UTC().Add(-olderThan)

	const query = `DELETE FROM quote_history WHERE recorded_at < ?`
	if _, err := r.db.ExecContext(ctx, query, cutoff); err != nil {
		return fmt.Errorf("db.PruneHistory: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) Close() error {
	if r == nil || r.db == nil {
		return nil
	}

	return r.db.Close()
}

func expandDBPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("empty database path")
	}

	if path == ":memory:" || strings.HasPrefix(path, "file:") {
		return path, nil
	}

	if path == "~" || strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		if path == "~" {
			return homeDir, nil
		}

		return filepath.Join(homeDir, path[2:]), nil
	}

	return path, nil
}

func ensureParentDir(path string) error {
	if path == ":memory:" || strings.HasPrefix(path, "file:") {
		return nil
	}

	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}

	return os.MkdirAll(dir, 0o755)
}
