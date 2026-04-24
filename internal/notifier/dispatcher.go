package notifier

import (
	"context"
	"errors"

	"github.com/yourname/tkr/pkg/models"
)

// Dispatcher fans out alert events to multiple notifiers.
// TODO(M6-5): implement concurrent notifier dispatch.
type Dispatcher struct {
	notifiers []Notifier
}

func NewDispatcher(notifiers []Notifier) *Dispatcher {
	// TODO(M6-5): initialize dispatcher.
	return &Dispatcher{notifiers: notifiers}
}

func (d *Dispatcher) Dispatch(ctx context.Context, event models.AlertEvent) error {
	_, _ = ctx, event
	// TODO(M6-5): implement dispatch and error aggregation.
	return errors.New("not implemented")
}
