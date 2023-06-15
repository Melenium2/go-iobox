package backoff

import (
	"math"
	"math/rand"
	"time"
)

const (
	DefaultMinDuration = 5 * time.Second
	DefaultMaxDuration = 60 * time.Second
	DefaultFactor      = float64(2)
)

type Config struct {
	Min    time.Duration
	Max    time.Duration
	Factor float64
}

func defaultConfig() Config {
	return Config{
		Min:    DefaultMinDuration,
		Max:    DefaultMaxDuration,
		Factor: DefaultFactor,
	}
}

type Backoff struct {
	c Config
}

func NewBackoff(cfg ...Config) *Backoff {
	c := defaultConfig()

	if len(cfg) > 0 {
		c = cfg[0]
	}

	if c.Max <= 0 {
		c.Max = DefaultMaxDuration
	}

	if c.Min <= 0 {
		c.Min = DefaultMinDuration
	}

	if c.Factor <= 0 {
		c.Factor = DefaultFactor
	}

	return &Backoff{
		c: c,
	}
}

func (b *Backoff) Next(attempt int) time.Duration {
	if attempt < 0 {
		return b.c.Max
	}

	if attempt == 0 {
		return b.c.Min
	}

	minf := float64(b.c.Min)
	attemptf := float64(attempt)
	durf := minf * math.Pow(b.c.Factor, attemptf)
	durf = rand.Float64()*(durf-minf) + minf

	dur := time.Duration(durf)

	if dur < b.c.Min {
		return b.c.Min
	}

	if dur > b.c.Max {
		return b.c.Max
	}

	return dur
}
