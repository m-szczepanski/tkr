package provider

import (
	"context"
	"errors"

	"github.com/yourname/tkr/pkg/models"
)

// Router routes quote/history requests across providers.
// TODO(M2-8): implement fallback, cache, and circuit breaker behavior.
type Router struct {
	providers []Provider
}

// NewRouter creates a provider router.
func NewRouter(providers []Provider) *Router {
	// TODO(M2-8): initialize router state.
	return &Router{providers: providers}
}

// DefaultProviders builds the default provider chain.
func DefaultProviders() []Provider {
	// TODO(M2-8): construct providers from config.
	return nil
}

// Quote fetches a quote through the provider chain.
func (r *Router) Quote(ctx context.Context, ticker string, exchange models.ExchangeID) (models.Quote, error) {
	_, _, _ = ctx, ticker, exchange
	// TODO(M2-8): implement provider routing and fallback.
	return models.Quote{}, errors.New("not implemented")
}

// History fetches historical OHLCV through the provider chain.
func (r *Router) History(ctx context.Context, ticker string, exchange models.ExchangeID, days int) ([]models.OHLCV, error) {
	_, _, _, _ = ctx, ticker, exchange, days
	// TODO(M2-8): implement provider routing for history.
	return nil, errors.New("not implemented")
}
