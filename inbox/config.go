package inbox

import (
	"time"

	"github.com/google/uuid"

	"github.com/Melenium2/go-iobox/retention"
)

const (
	// DefaultIterationRate is the timeout after which all events
	// in the inbox table will be processed.
	DefaultIterationRate = 5 * time.Second
	// DefaultIterationSeed is a number that is used to generate a random
	// duration for the next worker iteration.
	DefaultIterationSeed = 2
	// DefaultHandlerTimeout is the timeout after which the handler
	// will be stopped and the status will be set as Fail.
	DefaultHandlerTimeout = 10 * time.Second
	// DefaultRetryAttempts is the max attempts before event marks
	// as 'dead'. 'Dead' means that the event will no longer be
	// processed.
	DefaultRetryAttempts = 5
)

type (
	// DeadCallback prototype of function that is called if message is 'dead'
	DeadCallback func(eventID uuid.UUID, msg string)
	// ErrorCallback prototype of function that is called if errors occurs
	// during inbox process.
	ErrorCallback func(err error)
)

func nopDeadCallback(uuid.UUID, string) {}
func nopErrorCallback(err error)        {}

type config struct {
	iterationRate    time.Duration
	iterationSeed    int
	handlerTimeout   time.Duration
	maxRetryAttempts int
	retention        retention.Config
	onDead           DeadCallback
	onError          ErrorCallback
}

func defaultConfig() config {
	return config{
		iterationRate:    DefaultIterationRate,
		iterationSeed:    DefaultIterationSeed,
		handlerTimeout:   DefaultHandlerTimeout,
		maxRetryAttempts: DefaultRetryAttempts,
		retention:        retention.Config{},
		onDead:           nopDeadCallback,
		onError:          nopErrorCallback,
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

// WithIterationSeed sets the seed value for generating a random
// duration to add to DefaultIterationRate.
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

// WithMaxRetryAttempt sets custom max attempts for processing event.
func WithMaxRetryAttempt(maxAttempt int) Option {
	return func(c config) config {
		c.maxRetryAttempts = maxAttempt

		return c
	}
}

// WithRetention sets the retention configuration for inbox table.
//
// Arguments:
//
//	eraseInterval - interval for the next erase execution.
//	windowDays - the data older than the specified number of days will be deleted.
func WithRetention(eraseInterval time.Duration, windowDays int) Option {
	return func(c config) config {
		currCfg := c.retention
		currCfg.EraseInterval = eraseInterval
		currCfg.RetentionWindowDays = windowDays

		c.retention = currCfg

		return c
	}
}

// OnDeadCallback sets custom callback for each message that can not
// be processed and marks as 'dead'. Function fires if 'dead' message
// detected.
func OnDeadCallback(callback DeadCallback) Option {
	return func(c config) config {
		c.onDead = callback

		return c
	}
}

// ErrorCallback sets custom callback that is called if errors occurs
// during inbox process.
func OnErrorCallback(callback ErrorCallback) Option {
	return func(c config) config {
		c.onError = callback
		c.retention.ErrorCallback = callback

		return c
	}
}
