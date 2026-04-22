package notifier

import (
	"context"
	"errors"

	"github.com/yourname/tkr/pkg/models"
)

// WebhookNotifier dispatches alerts via HTTP webhooks.
// TODO(M6-4): implement webhook notifier.
type WebhookNotifier struct{}

func NewWebhookNotifier() *WebhookNotifier {
	// TODO(M6-4): initialize webhook notifier.
	return &WebhookNotifier{}
}

func (n *WebhookNotifier) ID() string {
	// TODO(M6-4): return channel ID.
	return "webhook"
}

func (n *WebhookNotifier) Send(ctx context.Context, event models.AlertEvent) error {
	_, _ = ctx, event
	// TODO(M6-4): implement webhook notification dispatch.
	return errors.New("not implemented")
}
