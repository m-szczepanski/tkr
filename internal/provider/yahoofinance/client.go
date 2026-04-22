package yahoofinance

import (
	"context"
	"errors"

	"github.com/yourname/tkr/pkg/models"
)

// Client is the Yahoo Finance provider implementation.
// TODO(M2-6): implement Yahoo Finance API integration.
type Client struct{}

func New() *Client {
	// TODO(M2-6): initialize Yahoo Finance client.
	return &Client{}
}

func (c *Client) ID() string {
	// TODO(M2-6): return provider ID.
	return "yahoofinance"
}

func (c *Client) Supports(exchange models.ExchangeID) bool {
	_ = exchange
	// TODO(M2-6): implement exchange support matrix.
	return false
}

func (c *Client) Quote(ctx context.Context, ticker string) (models.Quote, error) {
	_, _ = ctx, ticker
	// TODO(M2-6): implement quote fetch.
	return models.Quote{}, errors.New("not implemented")
}

func (c *Client) History(ctx context.Context, ticker string, days int) ([]models.OHLCV, error) {
	_, _, _ = ctx, ticker, days
	// TODO(M2-6): implement history fetch.
	return nil, errors.New("not implemented")
}
