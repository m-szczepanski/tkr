package provider

import (
	"context"

	"github.com/yourname/tkr/pkg/models"
)

// Provider is the interface all market data providers must implement.
// See .agent/AI_AGENT_GUIDE.md §2.1 for the full contract.
type Provider interface {
	ID() string
	Supports(exchange models.ExchangeID) bool
	Quote(ctx context.Context, ticker string) (models.Quote, error)
	History(ctx context.Context, ticker string, days int) ([]models.OHLCV, error)
}
