package outbox

import (
	"log"
	"time"

	"github.com/Melenium2/go-iobox/retention"
)

const tableName = "__outbox_table"

const (
	// DefaultIterationRate is the timeout after which all outbox events
	// in the outbox table are sent to the broker.
	//
	// Default: 5 * time.Second.
	DefaultIterationRate = 5 * time.Second
	// DefaultIterationSeed is a number that is used to generate a random
	// duration for the next worker iteration.
	//
	// Default: 2.
	DefaultIterationSeed = 2
	// DefaultPublishTimeout is the timeout after which the publication
	// of the current event is canceled. The current event marked as 'not yet published', and
	// processing continues.
	//
	// Default: 2 * time.Second.
	DefaultPublishTimeout = 2 * time.Second
	// DebugMode enables additional logs for debug outbox process.
	// Now, this option do nothing.
	//
	// Default: false.
	DebugMode = false
)

// TODO: Что то сделать с логером.
var DefaultLogger = log.Default()

type config struct {
	iterationRate time.Duration
	iterationSeed int
	timeout       time.Duration
	retention     retention.Config
	logger        Logger
	debugMode     bool
}

func defaultConfig() config {
	return config{
		iterationRate: DefaultIterationRate,
		iterationSeed: DefaultIterationSeed,
		timeout:       DefaultPublishTimeout,
		retention:     retention.Config{},
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

// WithIterationSeed sets the seed value for generating a random
// duration to add to DefaultIterationRate.
func WithIterationSeed(seed int) Option {
	return func(c config) config {
		c.iterationSeed = seed

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

// WithPublishTimeout sets a custom timeout for publishing next event.
func WithPublishTimeout(dur time.Duration) Option {
	return func(c config) config {
		c.timeout = dur

		return c
	}
}

// WithRetention sets the retention configuration for outbox table.
func WithRetention(cfg retention.Config) Option {
	return func(c config) config {
		c.retention = cfg

		return c
	}
}

func EnableDebugMode() Option {
	return func(c config) config {
		c.debugMode = true

		return c
	}
}
