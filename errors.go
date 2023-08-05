package storage

import "errors"

var (
	// ErrTimeProviderUnavailable TimeProvider无法提供时间
	ErrTimeProviderUnavailable = errors.New("time provider unavailable")
)
