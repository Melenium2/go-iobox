package inbox

import (
	"log"
	"time"
)

const (
	// DefaultIterationRate is the timeout after which all events
	// in the inbox table will be processed.
	//
	// Default: 5 * time.Second.
	DefaultIterationRate = 5 * time.Second
	// DefaultHandlerTimeout is the timeout after which the handler
	// will be stopped and the status will be set as Fail.
	//
	// Default: 10 * time.Second.
	DefaultHandlerTimeout = 10 * time.Second
	// DebugMode enables additional logs for debug inbox process.
	// Now, this option do nothing.
	//
	// Default: false.
	DebugMode = false
)

var DefaultLogger = log.Default()

type config struct {
	iterationRate  time.Duration
	handlerTimeout time.Duration
	logger         Logger
	debugMode      bool
}

func defaultConfig() config {
	return config{
		iterationRate:  DefaultIterationRate,
		handlerTimeout: DefaultHandlerTimeout,
		logger:         DefaultLogger,
		debugMode:      DebugMode,
	}
}

// Option sets specific configuration to the Inbox.
type Option func(config) config

// WithIterationRate sets new interval for process all inbox events.
func WithIterationRate(dur time.Duration) Option {
	return func(c config) config {
		c.iterationRate = dur

		return c
	}
}

// WithHandlerTimeout sets new interval after which handler will be stopped.
func WithHandlerTimeout(dur time.Duration) Option {
	return func(c config) config {
		c.handlerTimeout = dur

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
