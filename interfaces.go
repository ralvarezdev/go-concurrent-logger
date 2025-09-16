package go_concurrent_logger

import (
	"context"
)

type (
	// LoggerProducer is an interface for logging messages with different severity levels
	LoggerProducer interface {
		Log(content string, category Category)
		Info(content string)
		Error(err error)
		Warning(content string)
		Debug(content string)
		Close()
		IsClosed() bool
		Tag() string
		IsDebug() bool
	}

	// Logger is an interface for writing log messages to a file
	Logger interface {
		NewProducer(
			tag string,
		) (LoggerProducer, error)
		ChangeFilePath(newPath string) error
		Run(ctx context.Context, stopFn func()) error
		IsRunning() bool
		IsClosed() bool
		IsDebug() bool
		WaitUntilReady(ctx context.Context) error
	}
)
