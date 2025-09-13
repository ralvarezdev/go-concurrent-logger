package go_concurrent_logger

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