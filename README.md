# go-concurrent-logger

**Note:** This repository is archived and read-only.

A concurrent file logger for Go. Goroutines obtain lightweight producers that send tagged, categorised messages over a buffered channel to a single logger goroutine, which writes them to a log file.

## Installation

```bash
go get github.com/ralvarezdev/go-concurrent-logger
```

The package name is `go_concurrent_logger`. It depends on `go-context` and `go-crypto` by the same author.

## Usage

```go
import (
	"context"
	"time"

	gocl "github.com/ralvarezdev/go-concurrent-logger"
)

ctx, cancel := context.WithCancel(context.Background())
defer cancel()

logger, err := gocl.NewDefaultLogger(
	"logs/app.log", // file path (parent directory is created)
	5*time.Second,  // graceful shutdown timeout
	time.RFC3339,   // timestamp format
	100,            // channel buffer size
	4096,           // file buffer size
	"app",          // logger tag
)
if err != nil {
	panic(err)
}

// Run blocks, so start it in a goroutine
go func() { _ = logger.Run(ctx, cancel) }()
if err = logger.WaitUntilReady(ctx); err != nil {
	panic(err)
}

producer, err := logger.NewProducer("worker-1", true) // tag, debug enabled
if err != nil {
	panic(err)
}
defer producer.Close()

producer.Info("started")
producer.Debug("only written when debug is true")
```

Messages are formatted as `<timestamp> <CATEGORY> [<tag>] <content>`. Log files are created with mode `0644` and directories with `0755`.

## API

- **Categories** — `CategoryInfo`, `CategoryWarning`, `CategoryError`, `CategoryDebug`.
- **`Logger`** — `NewProducer(tag, debug)`, `ChangeFilePath(path)`, `Run(ctx, stopFn)`, `IsRunning()`, `IsClosed()`, `WaitUntilReady(ctx)`.
- **`LoggerProducer`** — `Log`, `Info`, `Error`, `Warning`, `Debug`, `Close`, `IsClosed`, `Tag`, `IsDebug`.
- **Implementations** — `DefaultLogger` and `DefaultLoggerProducer`; an empty producer tag is replaced by a random UUID.
- **Helpers** — `LogOnError(fn, producer)` and `CancelContextAndLogOnError(ctx, cancelFn, fn, producer)`.
- **Errors** — `ErrEmptyFilePath`, `ErrInvalidChannelBufferSize`, `ErrInvalidFileBufferSize`, `ErrEmptyTag`, `ErrLoggerAlreadyRunning`, `ErrLoggerNotRunning`, `ErrLoggerClosed`, `ErrNilLogger`, `ErrNilSendFunction`, `ErrNilCloseFunction`.

There are no tests.

## License

GNU General Public License v3.0. See [LICENSE](LICENSE).
