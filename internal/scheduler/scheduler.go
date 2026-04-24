package scheduler

import (
	"context"
	"errors"
)

// Scheduler runs periodic quote polling and alert evaluation.
// TODO(M7-1): implement scheduler lifecycle and polling loop.
type Scheduler struct{}

func New() *Scheduler {
	// TODO(M7-1): initialize scheduler dependencies.
	return &Scheduler{}
}

func (s *Scheduler) Start(ctx context.Context) error {
	_ = ctx
	// TODO(M7-1): implement scheduler start.
	return errors.New("not implemented")
}

func (s *Scheduler) Stop() error {
	// TODO(M7-1): implement scheduler stop.
	return errors.New("not implemented")
}
