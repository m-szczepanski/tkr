package apperrors

import "errors"

var (
	ErrTickerNotFound   = errors.New("ticker not found")
	ErrNoProvider       = errors.New("no provider available")
	ErrRateLimited      = errors.New("rate limited")
	ErrDB               = errors.New("database error")
	ErrConfig           = errors.New("configuration error")
	ErrDaemonRunning    = errors.New("daemon already running")
	ErrDaemonNotRunning = errors.New("daemon not running")
	ErrNotifyFail       = errors.New("notification failed")
	ErrInvalidCondition = errors.New("invalid condition expression")
)
