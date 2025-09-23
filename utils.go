package go_concurrent_logger

import (
	"context"
	"fmt"

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

// CancelContextAndLogOnError cancels the context if an error is encountered and logs the error.
//
// Parameters:
//
// ctx: The context to be stopped.
// cancelFn: A function that cancels the context with a given error.
// fn: A function that returns an error.
// loggerProducer: The logger producer to log messages.
//
// Returns:
//
// A function that executes the provided function and cancels the context if an error occurs.
func CancelContextAndLogOnError(
	ctx context.Context,
	cancelFn context.CancelFunc,
	fn func(context.Context) error,
	loggerProducer LoggerProducer,
) func() error {
	return func() error {
		// Recover from panic and log it to prevent unregistered goroutine panics
		defer func() {
			if r := recover(); r != nil {
				if loggerProducer != nil {
					loggerProducer.Error(
						fmt.Errorf("goroutine panicked: %v", r),
					)
				} else {
					fmt.Println("Goroutine panicked:", r)
				}
			}
		}()

		// Cancel the context on error and log the error if it occurs
		err := gocontext.CancelContextOnError(
			ctx,
			cancelFn,
			fn,
		)()
		if err != nil && loggerProducer != nil {
			loggerProducer.Error(err)
		}
		return err
	}
}
