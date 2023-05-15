package outbox

import (
	"log"
	"time"
)

const (
	// DefaultIterationRate is a timeout after which all outbox events
	// in the outbox table are sent to the broker.
	//
	// Default: 5 * time.Second.
	DefaultIterationRate = 5 * time.Second
	// DebugMode enables additional logs for debug outbox process.
	// Now, this option do nothing.
	//
	// Default: false.
	DebugMode = false
)

var DefaultLogger = log.Default()

type config struct {
	iterationRate time.Duration
	logger        Logger
	debugMode     bool
}

func defaultConfig() config {
	return config{
		iterationRate: DefaultIterationRate,
		logger:        DefaultLogger,
		debugMode:     DebugMode,
	}
}

// Option sets specific configuration to the Outbox.
type Option func(config) config

// WithIterationRate sets new interval for sending events from the
// outbox table.
func WithIterationRate(dur time.Duration) Option {
	return func(c config) config {
		c.iterationRate = dur

		return c
	}
}

// WithLogger sets custom implementation of Logger.
func WithLogger(logger Logger) Option {
	return func(c config) config {
		c.logger = logger

		return c
	}
}

func EnableDebugMode() Option {
	return func(c config) config {
		c.debugMode = true

		return c
	}
}
