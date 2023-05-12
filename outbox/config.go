package outbox

import (
	"log"
	"time"
)

const (
	DefaultIterationRate = 5 * time.Second
	DebugMode            = false
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

type Option func(config) config

func WithIterationRate(dur time.Duration) Option {
	return func(c config) config {
		c.iterationRate = dur

		return c
	}
}

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
