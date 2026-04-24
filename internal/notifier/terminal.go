package notifier

import (
	"context"
	"errors"

	"github.com/yourname/tkr/pkg/models"
)

// TerminalNotifier dispatches alerts to terminal/desktop notifications.
// TODO(M6-1): implement terminal notifier.
type TerminalNotifier struct{}

func NewTerminalNotifier() *TerminalNotifier {
	// TODO(M6-1): initialize terminal notifier.
	return &TerminalNotifier{}
}

func (n *TerminalNotifier) ID() string {
	// TODO(M6-1): return channel ID.
	return "terminal"
}

func (n *TerminalNotifier) Send(ctx context.Context, event models.AlertEvent) error {
	_, _ = ctx, event
	// TODO(M6-1): implement terminal notification dispatch.
	return errors.New("not implemented")
}
