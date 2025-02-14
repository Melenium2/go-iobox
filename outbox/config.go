package outbox

import (
	"time"

	"github.com/Melenium2/go-iobox/retention"
)

const (
	// DefaultIterationRate is the timeout after which all outbox events
	// in the outbox table are sent to the broker.
	DefaultIterationRate = 5 * time.Second
	// DefaultIterationSeed is a number that is used to generate a random
	// duration for the next worker iteration.
	DefaultIterationSeed = 2
	// DefaultPublishTimeout is the timeout after which the publication
	// of the current event is canceled. The current event marked as 'not yet published', and
	// processing continues.
	DefaultPublishTimeout = 2 * time.Second
)

// ErrorCallback prototype of function that is called if errors occurs
// during outbox process.
type ErrorCallback func(err error)

func nopCallback(error) {}

type config struct {
	iterationRate time.Duration
	iterationSeed int
	timeout       time.Duration
	retention     retention.Config
	onError       ErrorCallback
}

func defaultConfig() config {
	return config{
		iterationRate: DefaultIterationRate,
		iterationSeed: DefaultIterationSeed,
		timeout:       DefaultPublishTimeout,
		retention:     retention.Config{},
		onError:       nopCallback,
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

// WithIterationSeed sets the seed value for generating a random
// duration to add to DefaultIterationRate.
func WithIterationSeed(seed int) Option {
	return func(c config) config {
		c.iterationSeed = seed

		return c
	}
}

// WithPublishTimeout sets a custom timeout for publishing next event.
func WithPublishTimeout(dur time.Duration) Option {
	return func(c config) config {
		c.timeout = dur

		return c
	}
}

// WithRetention sets the retention configuration for outbox table.
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

// ErrorCallback sets custom callback that is called if errors occurs
// during outbox process.
func OnErrorCallback(callback ErrorCallback) Option {
	return func(c config) config {
		c.onError = callback
		c.retention.ErrorCallback = callback

		return c
	}
}
