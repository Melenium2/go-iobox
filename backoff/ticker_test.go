package backoff

import (
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStopTicker(t *testing.T) {
	t.Run("should close stop channel", func(t *testing.T) {
		ticker := &Ticker{
			stop: make(chan struct{}),
		}

		ticker.stopTicker()

		_, ok := <-ticker.stop
		assert.False(t, ok)
	})
}

func TestStartTimer(t *testing.T) {
	t.Run("should start new timer and wait for fire", func(t *testing.T) {
		ticker := &Ticker{}

		ch := ticker.startTimer(100 * time.Millisecond)

		tick := <-ch
		assert.NotEmpty(t, tick)
	})
}

func TestBackoffDuration(t *testing.T) {
	t.Run("should calculate next duration with the custom backoff factor and base duration", func(t *testing.T) {
		baseDuration := 3 * time.Second

		ticker := &Ticker{
			backoff: NewBackoff(),
		}

		dur := ticker.backoffDuration(baseDuration, 1)
		dur2 := ticker.backoffDuration(baseDuration, 2)
		assert.Greater(t, dur2, dur)
	})
}

func TestNext(t *testing.T) {
	t.Run("should return new channel and wait for fire", func(t *testing.T) {
		var (
			ch           = make(chan time.Time)
			baseDuration = time.Duration(0)
			factor       = 2
			cfg          = Config{
				Min: baseDuration,
				Max: baseDuration + time.Second,
			}
		)

		ticker := &Ticker{
			C:            ch,
			c:            ch,
			baseDuration: baseDuration,
			backoffSeed:  factor,
			backoff:      NewBackoff(cfg),
		}

		now := time.Now()

		go func() {
			// NOTE: Next function is blocked until we read first value.
			for range ticker.C {
			}
		}()

		resCh := ticker.next(now)

		tick := <-resCh
		assert.Greater(t, tick, now)
	})

	t.Run("should return nil if stop channel is closed", func(t *testing.T) {
		ch := make(chan time.Time)

		ticker := &Ticker{
			C:       ch,
			c:       ch,
			backoff: NewBackoff(),
			stop:    make(chan struct{}),
		}

		close(ticker.stop)

		resCh := ticker.next(time.Now())
		assert.Nil(t, resCh)
	})
}

func TestRun(t *testing.T) {
	t.Run("should tick with specific backoff until stop", func(t *testing.T) {
		backoffConfig := Config{
			Min:    1 * time.Second,
			Max:    5 * time.Second,
			Factor: 2,
		}

		backoff := NewBackoff(backoffConfig)

		ticker := NewTicker(backoff, 5*time.Second, 2)

		go func() {
			time.Sleep(3 * time.Second)
			ticker.Stop()
		}()

		ticks := 0
		for tick := range ticker.C {
			slog.Info(tick.String())
			ticks++
		}

		assert.Greater(t, ticks, 0)
	})
}
