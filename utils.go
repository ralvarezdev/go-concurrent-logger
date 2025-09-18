package go_concurrent_logger

import (
	"context"

	gocontext "github.com/ralvarezdev/go-context"
)

// LogOnError logs the error if it's not nil.
//
// Parameters:
//
// fn: The function to execute that returns an error.
// loggerProducer: The logger producer to use for logging the error.
func LogOnError(fn func() error, loggerProducer LoggerProducer) error {
	// If the logger producer is nil, just execute the function
	if loggerProducer == nil {
		return fn()
	}

	// Execute the function and log the error if it's not nil
	err := fn()
	if err != nil {
		loggerProducer.Error(err)
	}
	return err
}

// StopContextAndLogOnError stops the context if an error is encountered and logs the error.
//
// Parameters:
//
// ctx: The context to be stopped.
// stopFn: A function that stops the context with a given error.
// fn: A function that returns an error.
// loggerProducer: The logger producer to log messages.
//
// Returns:
//
// A function that executes the provided function and stops the context if an error occurs.
func StopContextAndLogOnError(
	ctx context.Context,
	stopFn func(),
	fn func(context.Context) error,
	loggerProducer LoggerProducer,
) func() error {
	return func() error {
		err := gocontext.StopContextOnError(
			ctx,
			stopFn,
			fn,
		)()
		if err != nil && loggerProducer != nil {
			loggerProducer.Error(err)
		}
		return err
	}
}
