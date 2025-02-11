package inbox

import (
	"log"
	"time"

	"github.com/google/uuid"
)

const (
	// DefaultIterationRate is the timeout after which all events
	// in the inbox table will be processed.
	//
	// Default: 5 * time.Second.
	DefaultIterationRate = 5 * time.Second
	// DefaultIterationSeed is a number that is used to generate a random
	// duration for the next worker iteration.
	//
	// Default: 2.
	DefaultIterationSeed = 2
	// DefaultHandlerTimeout is the timeout after which the handler
	// will be stopped and the status will be set as Fail.
	//
	// Default: 10 * time.Second.
	DefaultHandlerTimeout = 10 * time.Second
	// DefaultRetryAttempts is the max attempts before event marks
	// as 'dead'. 'Dead' means that the event will no longer be
	// processed.
	//
	// Default: 5.
	DefaultRetryAttempts = 5
	// DebugMode enables additional logs for debug inbox process.
	// Now, this option do nothing.
	//
	// Default: false.
	DebugMode = false
)

var DefaultLogger = log.Default()

// NopLogger logs nothing. Use it if you want
// mute Inbox.
type NopLogger struct{}

func NewNopLogger() *NopLogger { return &NopLogger{} }

func (l *NopLogger) Print(...any)          {}
func (l *NopLogger) Printf(string, ...any) {}

// ErrorCallback prototype of function that can be called on failed or
// dead message.
type ErrorCallback func(eventID uuid.UUID, msg string)

func emptyCallback(uuid.UUID, string) {}

type config struct {
	iterationRate    time.Duration
	iterationSeed    int
	handlerTimeout   time.Duration
	maxRetryAttempts int
	logger           Logger
	debugMode        bool
	onDeadCallback   ErrorCallback
}

func defaultConfig() config {
	return config{
		iterationRate:    DefaultIterationRate,
		iterationSeed:    DefaultIterationSeed,
		handlerTimeout:   DefaultHandlerTimeout,
		maxRetryAttempts: DefaultRetryAttempts,
		logger:           DefaultLogger,
		debugMode:        DebugMode,
		onDeadCallback:   emptyCallback,
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

func WithIterationSeed(seed int) Option {
	return func(c config) config {
		c.iterationSeed = seed

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

// WithMaxRetryAttempt sets custom max attempts for processing event.
func WithMaxRetryAttempt(maxAttempt int) Option {
	return func(c config) config {
		c.maxRetryAttempts = maxAttempt

		return c
	}
}

func EnableDebugMode() Option {
	return func(c config) config {
		c.debugMode = true

		return c
	}
}

// OnDeadCallback sets custom callback for each message that can not
// be processed and marks as 'dead'. Function fires if 'dead' message
// detected.
func OnDeadCallback(callback ErrorCallback) Option {
	return func(c config) config {
		c.onDeadCallback = callback

		return c
	}
}
