package notifier

import (
	"context"

	"github.com/yourname/tkr/pkg/models"
)

// Notifier represents a notification channel implementation.
// TODO(M6-5): implement channel wiring and concrete notifier registration.
type Notifier interface {
	ID() string
	Send(ctx context.Context, event models.AlertEvent) error
}
