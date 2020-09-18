package errors

import (
	"errors"
)

func IsRateLimit(err error) bool {
	return (Cause(err) == ErrRateLimit)
}

func IsCircuitBreak(err error) bool {
	return (Cause(err) == ErrCircuitBreak)
}

var (
	ErrRateLimit    = errors.New("rate limit triggered")
	ErrCircuitBreak = errors.New("circuit break triggered")
)
