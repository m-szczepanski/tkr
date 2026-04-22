package notifier

import (
	"context"
	"errors"

	"github.com/yourname/tkr/pkg/models"
)

// EmailNotifier dispatches alerts via SMTP email.
// TODO(M6-2): implement email notifier.
type EmailNotifier struct{}

func NewEmailNotifier() *EmailNotifier {
	// TODO(M6-2): initialize email notifier.
	return &EmailNotifier{}
}

func (n *EmailNotifier) ID() string {
	// TODO(M6-2): return channel ID.
	return "email"
}

func (n *EmailNotifier) Send(ctx context.Context, event models.AlertEvent) error {
	_, _ = ctx, event
	// TODO(M6-2): implement email notification dispatch.
	return errors.New("not implemented")
}
