package db

import (
	"context"
	"errors"
	"time"

	"github.com/yourname/tkr/pkg/models"
)

// SQLiteRepository is the SQLite-backed Repository implementation.
// TODO(M1-6): implement SQL-backed methods.
type SQLiteRepository struct{}

// Open opens a SQLite repository at the given path.
func Open(path string) (Repository, error) {
	_ = path
	// TODO(M1-6): initialize sqlite connection and pragmas.
	return &SQLiteRepository{}, nil
}

func (r *SQLiteRepository) AddStock(ctx context.Context, s models.Stock) error {
	_, _ = ctx, s
	// TODO(M3-1): implement AddStock.
	return errors.New("not implemented")
}

func (r *SQLiteRepository) RemoveStock(ctx context.Context, ticker string) error {
	_, _ = ctx, ticker
	// TODO(M3-1): implement RemoveStock.
	return errors.New("not implemented")
}

func (r *SQLiteRepository) ListStocks(ctx context.Context) ([]models.Stock, error) {
	_ = ctx
	// TODO(M3-1): implement ListStocks.
	return nil, errors.New("not implemented")
}

func (r *SQLiteRepository) GetStock(ctx context.Context, ticker string) (models.Stock, error) {
	_, _ = ctx, ticker
	// TODO(M3-1): implement GetStock.
	return models.Stock{}, errors.New("not implemented")
}

func (r *SQLiteRepository) AddAlertRule(ctx context.Context, rule models.AlertRule) (int64, error) {
	_, _ = ctx, rule
	// TODO(M4-3): implement AddAlertRule.
	return 0, errors.New("not implemented")
}

func (r *SQLiteRepository) RemoveAlertRule(ctx context.Context, id int64) error {
	_, _ = ctx, id
	// TODO(M4-3): implement RemoveAlertRule.
	return errors.New("not implemented")
}

func (r *SQLiteRepository) ListAlertRules(ctx context.Context, ticker string) ([]models.AlertRule, error) {
	_, _ = ctx, ticker
	// TODO(M4-3): implement ListAlertRules.
	return nil, errors.New("not implemented")
}

func (r *SQLiteRepository) UpdateAlertRule(ctx context.Context, rule models.AlertRule) error {
	_, _ = ctx, rule
	// TODO(M4-3): implement UpdateAlertRule.
	return errors.New("not implemented")
}

func (r *SQLiteRepository) AddAlertEvent(ctx context.Context, event models.AlertEvent) error {
	_, _ = ctx, event
	// TODO(M4-3): implement AddAlertEvent.
	return errors.New("not implemented")
}

func (r *SQLiteRepository) ListAlertEvents(ctx context.Context, filter models.AlertEventFilter) ([]models.AlertEvent, error) {
	_, _ = ctx, filter
	// TODO(M4-3): implement ListAlertEvents.
	return nil, errors.New("not implemented")
}

func (r *SQLiteRepository) SaveQuote(ctx context.Context, q models.Quote) error {
	_, _ = ctx, q
	// TODO(M1-6): implement SaveQuote.
	return errors.New("not implemented")
}

func (r *SQLiteRepository) GetRecentQuotes(ctx context.Context, ticker string, limit int) ([]models.Quote, error) {
	_, _, _ = ctx, ticker, limit
	// TODO(M1-6): implement GetRecentQuotes.
	return nil, errors.New("not implemented")
}

func (r *SQLiteRepository) PruneHistory(ctx context.Context, olderThan time.Duration) error {
	_, _ = ctx, olderThan
	// TODO(M1-6): implement PruneHistory.
	return errors.New("not implemented")
}
