package go_concurrent_logger

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	gocontext "github.com/ralvarezdev/go-context"
	gocryptouuid "github.com/ralvarezdev/go-crypto/uuid"
)

type (
	// DefaultLoggerProducer is the default LoggerProducer implementation for messages with different severity levels
	DefaultLoggerProducer struct {
		sendFn          func(*Message)
		closed          atomic.Bool
		closeFn         func()
		tag             string
		mutex           sync.Mutex
		timestampFormat string
		debug           bool
	}

	// DefaultLogger is the default Logger implementation to handle writing log messages to a file.
	DefaultLogger struct {
		messagesCh              chan *Message
		readyCh                 chan struct{}
		wgProducers             sync.WaitGroup
		closed                  atomic.Bool
		isRunning               atomic.Bool
		mutex                   sync.Mutex
		filePath                string
		gracefulShutdownTimeout time.Duration
		timestampFormat         string
		channelBufferSize       int
		fileBufferSize          int
		tag                     string
	}
)

// NewDefaultLoggerProducer creates a new DefaultLoggerProducer instance.
//
// Parameters:
//
// timestampFormat: Format for the timestamp.
// sendFn: Function to send messages.
// closeFn: Function to call when done.
// tag: Tag to identify the logger instance.
// debug: Flag to indicate if the logger is in debug mode.
//
// Returns:
//
// A pointer to a DefaultLoggerProducer instance or an error if the parameters are invalid.
func NewDefaultLoggerProducer(
	timestampFormat string,
	sendFn func(*Message),
	closeFn func(),
	tag string,
	debug bool,
) (*DefaultLoggerProducer, error) {
	// Check if the sendFn is nil
	if sendFn == nil {
		return nil, ErrNilSendFunction
	}

	// Check if the closeFn is nil
	if closeFn == nil {
		return nil, ErrNilCloseFunction
	}

	// Generate a unique tag if required
	if tag == "" {
		// Generate a UUID to ensure uniqueness
		uniqueID, err := gocryptouuid.NewUUIDv4()
		if err != nil {
			return nil, err
		}
		tag = uniqueID
	}

	// Create a new DefaultLoggerProducer instance
	producer := &DefaultLoggerProducer{
		timestampFormat: timestampFormat,
		sendFn:          sendFn,
		closeFn:         closeFn,
		tag:             tag,
		debug:           debug,
	}

	// Log the initialization if a tag is provided
	producer.Debug("Initializing new DefaultLoggerProducer")

	return producer, nil
}

// Log logs a message with the specified content and category.
//
// Parameters:
//
// content: The content of the log message.
// category: The category of the log message.
func (l *DefaultLoggerProducer) Log(content string, category Category) {
	// Create a message object
	message := NewMessage(category, content, l.tag, l.timestampFormat)

	// Send the message if the logger is not closed
	if l.IsClosed() {
		fmt.Println("Logger is closed. Cannot log message: ", message.String())
		return
	}
	l.sendFn(message)
}

// Info logs an informational message.
//
// Parameters:
//
// content: The content of the informational message.
func (l *DefaultLoggerProducer) Info(content string) {
	l.Log(content, CategoryInfo)
}

// Error logs an error message.
//
// Parameters:
//
// err: The error to log.
func (l *DefaultLoggerProducer) Error(err error) {
	l.Log(fmt.Sprintf("An error occurred: %v", err), CategoryError)
}

// Warning logs a warning message.
//
// Parameters:
//
// content: The content of the warning message.
func (l *DefaultLoggerProducer) Warning(content string) {
	l.Log(content, CategoryWarning)
}

// Debug logs a debug message if the logger is in debug mode.
//
// Parameters:
//
// content: The content of the debug message.
func (l *DefaultLoggerProducer) Debug(content string) {
	if l.debug {
		l.Log(content, CategoryDebug)
	}
}

// Close signals that the logger is done and performs cleanup.
func (l *DefaultLoggerProducer) Close() {
	l.mutex.Lock()

	// Check if already closed to prevent multiple calls
	if l.IsClosed() {
		l.mutex.Unlock()
		return
	}

	// Send a final debug message if in debug mode
	l.Debug("Closed")

	// Mark the logger as closed
	l.closed.Store(true)

	l.mutex.Unlock()

	// Call the close function to signal completion
	l.closeFn()
}

// IsClosed returns true if the logger producer has been closed.
//
// Returns:
//
// True if the logger producer is closed, otherwise false.
func (l *DefaultLoggerProducer) IsClosed() bool {
	return l.closed.Load()
}

// Tag returns the tag associated with the logger producer.
//
// Returns:
//
// The tag string.
func (l *DefaultLoggerProducer) Tag() string {
	return l.tag
}

// IsDebug returns true if the logger producer is in debug mode.
//
// Returns:
//
// True if the logger producer is in debug mode, otherwise false.
func (l *DefaultLoggerProducer) IsDebug() bool {
	return l.debug
}

// NewDefaultLogger creates a new DefaultLogger instance.
//
// Parameters:
//
// filePath: Path to the log file.
// gracefulShutdownTimeout: Duration to wait for graceful shutdown.
// timestampFormat: Format for timestamps in log messages.
// channelBufferSize: Size of the message channel buffer.
// fileBufferSize: Size of the file write buffer.
// tag: Default tag for log messages.
//
// Returns:
//
// A pointer to a DefaultLogger instance.
func NewDefaultLogger(
	filePath string,
	gracefulShutdownTimeout time.Duration,
	timestampFormat string,
	channelBufferSize int,
	fileBufferSize int,
	tag string,
) (*DefaultLogger, error) {
	// Validate the file path
	if filePath == "" {
		return nil, ErrEmptyFilePath
	}

	// Validate channel and file buffer sizes
	if channelBufferSize <= 0 {
		return nil, ErrInvalidChannelBufferSize
	}
	if fileBufferSize <= 0 {
		return nil, ErrInvalidFileBufferSize
	}
	if tag == "" {
		return nil, ErrEmptyTag
	}

	return &DefaultLogger{
		filePath:                filePath,
		gracefulShutdownTimeout: gracefulShutdownTimeout,
		timestampFormat:         timestampFormat,
		channelBufferSize:       channelBufferSize,
		fileBufferSize:          fileBufferSize,
		tag:                     tag,
		readyCh:                 make(chan struct{}),
	}, nil
}

// ChangeFilePath changes the log file path to a new path.
//
// Parameters:
//
// newPath: The new file path for the log file.
func (l *DefaultLogger) ChangeFilePath(newPath string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Validate the new file path
	if newPath == "" {
		return ErrEmptyFilePath
	}

	// Update the file path
	l.filePath = newPath
	return nil
}

// IsRunning returns true if the logger is currently running.
//
// Returns:
//
// True if the logger is running, otherwise false.
func (l *DefaultLogger) IsRunning() bool {
	return l.isRunning.Load()
}

// runToWrap is an internal function to process and write messages to the log file.
//
// Parameters:
//
// ctx: The context to control cancellation and timeouts.
//
// Returns:
//
// An error if any issues occur during message processing or file writing.
func (l *DefaultLogger) runToWrap(ctx context.Context) error {
	// Get the current file path
	l.mutex.Lock()
	currentFilePath := l.filePath
	l.mutex.Unlock()

	// Ensure parent directory exists
	logDir := filepath.Dir(currentFilePath)
	if err := os.MkdirAll(logDir, dirPerm); err != nil {
		return fmt.Errorf("creating log directory %s: %w", logDir, err)
	}

	// Open the log file in append mode
	file, err := os.OpenFile(
		currentFilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		filePerm,
	)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			fmt.Printf("Error closing file: %v\n", err)
		}
	}(file)

	// Create a buffered writer for efficient file writing
	buf := bufio.NewWriterSize(file, l.fileBufferSize)
	defer func(buf *bufio.Writer) {
		if err := buf.Flush(); err != nil {
			fmt.Printf("Error flushing buffer: %v\n", err)
		}
	}(buf)

	// Create a helper function to write a message to the buffer
	writeLine := func(m *Message) error {
		if m == nil {
			m = NewMessage(
				CategoryWarning,
				"Nil message received",
				l.tag,
				l.timestampFormat,
			)
		}
		if _, err := buf.WriteString(m.String() + "\n"); err != nil {
			return err
		}
		return nil
	}

	// Log a message indicating that the writer has started
	if err = writeLine(
		NewMessage(
			CategoryInfo,
			"Started",
			l.tag,
			l.timestampFormat,
		),
	); err != nil {
		return err
	}

	// Process messages from the channel
	for {
		select {
		case <-ctx.Done():
			// Log a message indicating that the context was cancelled
			_ = writeLine(
				NewMessage(
					CategoryInfo,
					"Context cancelled by caller",
					l.tag,
					l.timestampFormat,
				),
			)
			_ = buf.Flush()

			// Start a timer to continue receiving messages for a short period
			timer := time.NewTimer(l.gracefulShutdownTimeout)
			defer timer.Stop()
			for {
				select {
				case msg, ok := <-l.messagesCh:
					if !ok {
						return ctx.Err()
					}
					if err := writeLine(msg); err != nil {
						return err
					}
					_ = buf.Flush()
				case <-timer.C:
					_ = writeLine(
						NewMessage(
							CategoryInfo,
							"Grace period ended",
							l.tag,
							l.timestampFormat,
						),
					)
					_ = buf.Flush()
					return ctx.Err()
				}
			}
		case msg, ok := <-l.messagesCh:
			if !ok {
				// Channel is closed, exit the loop
				_ = writeLine(
					NewMessage(
						CategoryInfo,
						"Messages channel closed",
						l.tag,
						l.timestampFormat,
					),
				)
				_ = buf.Flush()
				return nil
			}
			if err = writeLine(msg); err != nil {
				return err
			}
			_ = buf.Flush()
		}
	}
}

// Run processes and writes all received messages to the log file until the channel is closed or the context is cancelled.
//
// Parameters:
//
// ctx: The context to control cancellation and timeouts.
// stopFn: Function to call when stopping the handler.
//
// Returns:
//
// An error if any issues occur during message processing or file writing.
func (l *DefaultLogger) Run(ctx context.Context, stopFn func()) error {
	l.mutex.Lock()

	// Check if it's already running
	if l.IsRunning() {
		l.mutex.Unlock()
		return ErrLoggerAlreadyRunning
	}
	defer func() {
		l.mutex.Lock()

		// Set running to false
		l.isRunning.Store(false)

		l.mutex.Unlock()
	}()

	// Set running to true
	l.isRunning.Store(true)

	// Reset the closed state
	l.closed.Store(false)

	// Reinitialize the messages channel
	l.messagesCh = make(chan *Message, l.channelBufferSize)
	close(l.readyCh)
	defer l.close()

	l.mutex.Unlock()

	return gocontext.StopContextOnError(
		ctx, stopFn, l.runToWrap,
	)()
}

// NewProducer returns a new LoggerProducer instance associated with this DefaultLogger.
//
// Parameters:
//
// tag: Tag to identify the logger instance.
// debug: Flag to indicate if the logger is in debug mode.
//
// Returns:
//
// A pointer to a LoggerProducer instance or an error if the parameters are invalid.
func (l *DefaultLogger) NewProducer(
	tag string,
	debug bool,
) (LoggerProducer, error) {
	l.mutex.Lock()

	// Check if the logger is already closed
	if l.IsClosed() {
		return nil, ErrLoggerClosed
	}

	// Check if the logger is running
	if !l.IsRunning() {
		l.mutex.Unlock()
		return nil, ErrLoggerNotRunning
	}

	// Increment the producer wait group counter
	l.wgProducers.Add(1)
	l.mutex.Unlock()

	// Create and return a new DefaultLoggerProducer instance
	loggerProducer, err := NewDefaultLoggerProducer(
		l.timestampFormat,
		func(m *Message) {
			l.messagesCh <- m
		},
		func() { l.wgProducers.Done() },
		tag,
		debug,
	)
	if err != nil {
		l.wgProducers.Done()
		return nil, err
	}
	return loggerProducer, nil
}

// close signals no more producers will send; safe to call multiple times.
func (l *DefaultLogger) close() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Check if the logger is already closed
	if l.IsClosed() {
		return
	}

	// Mark the logger as closed
	l.closed.Store(true)

	// Wait for all registered producers to finish, then close channel.
	l.wgProducers.Wait()

	// Close the messages channel to signal no more messages will be sent.
	close(l.messagesCh)
	l.messagesCh = nil

	// Create a new ready channel for future runs
	l.readyCh = make(chan struct{})

	// Set running to false
	l.isRunning.Store(false)

	// Reset the producer wait group
	l.wgProducers = sync.WaitGroup{}
}

// WaitUntilReady blocks until the logger is ready or the context is done.
//
// Parameters:
//
// ctx: The context to control cancellation and timeouts.
//
// Returns:
//
// An error if the context is done before the logger is ready.
func (l *DefaultLogger) WaitUntilReady(ctx context.Context) error {
	select {
	case <-l.readyCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// IsClosed returns true if the logger channel has been closed.
//
// Returns:
//
// True if the logger channel is closed, otherwise false.
func (l *DefaultLogger) IsClosed() bool {
	return l.closed.Load()
}
