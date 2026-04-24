package stooq

import (
	"context"
	"errors"

	"github.com/yourname/tkr/pkg/models"
)

// Client is the Stooq provider implementation.
// TODO(M2-4): implement Stooq CSV integration.
type Client struct{}

func New() *Client {
	// TODO(M2-4): initialize Stooq client.
	return &Client{}
}

func (c *Client) ID() string {
	// TODO(M2-4): return provider ID.
	return "stooq"
}

func (c *Client) Supports(exchange models.ExchangeID) bool {
	_ = exchange
	// TODO(M2-4): implement exchange support matrix.
	return false
}

func (c *Client) Quote(ctx context.Context, ticker string) (models.Quote, error) {
	_, _ = ctx, ticker
	// TODO(M2-4): implement quote fetch.
	return models.Quote{}, errors.New("not implemented")
}

func (c *Client) History(ctx context.Context, ticker string, days int) ([]models.OHLCV, error) {
	_, _, _ = ctx, ticker, days
	// TODO(M2-4): implement history fetch.
	return nil, errors.New("not implemented")
}
