package outbox

import "time"

const (
	DefaultIterationRate = 5 * time.Second
)

type config struct {
	iterationRate time.Duration
}

func defaultConfig() config {
	return config{
		iterationRate: DefaultIterationRate,
	}
}

type Option func(config) config

func WithIterationRate(dur time.Duration) Option {
	return func(c config) config {
		c.iterationRate = dur

		return c
	}
}
