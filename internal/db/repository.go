package db

import (
	"context"
	"time"

	"github.com/yourname/tkr/pkg/models"
)

// Repository defines all persistence operations.
// TODO(M1-6): implement sqlite-backed repository methods.
type Repository interface {
	// Stocks
	AddStock(ctx context.Context, s models.Stock) error
	RemoveStock(ctx context.Context, ticker string) error
	ListStocks(ctx context.Context) ([]models.Stock, error)
	GetStock(ctx context.Context, ticker string) (models.Stock, error)

	// Alert rules
	AddAlertRule(ctx context.Context, rule models.AlertRule) (int64, error)
	RemoveAlertRule(ctx context.Context, id int64) error
	ListAlertRules(ctx context.Context, ticker string) ([]models.AlertRule, error)
	UpdateAlertRule(ctx context.Context, rule models.AlertRule) error

	// Alert events
	AddAlertEvent(ctx context.Context, event models.AlertEvent) error
	ListAlertEvents(ctx context.Context, filter models.AlertEventFilter) ([]models.AlertEvent, error)

	// Quote history
	SaveQuote(ctx context.Context, q models.Quote) error
	GetRecentQuotes(ctx context.Context, ticker string, limit int) ([]models.Quote, error)
	PruneHistory(ctx context.Context, olderThan time.Duration) error
}
