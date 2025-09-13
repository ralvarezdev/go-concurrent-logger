package go_concurrent_logger

import (
	"errors"
)

var (
	ErrNilSendFunction      = errors.New("send function cannot be nil")
	ErrNilCloseFunction     = errors.New("close function cannot be nil")
	ErrLoggerClosed         = errors.New("logger is closed")
	ErrNilLogger            = errors.New("logger cannot be nil")
	ErrLoggerAlreadyRunning = errors.New("logger is already running")
	ErrEmptyFilePath       = errors.New("file path cannot be empty")
	ErrInvalidChannelBufferSize = errors.New("channel buffer size must be greater than zero")
	ErrInvalidFileBufferSize    = errors.New("file buffer size must be greater than zero")
	ErrEmptyTag                 = errors.New("tag cannot be empty")
)
