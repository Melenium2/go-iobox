package backoff

import (
	"math/rand/v2"
	"sync"
	"time"
)

// Random constant used to generate a random backoff factor.
const backoffRange = 1_000_000

// Ticker extends logic of standard time.Ticker with
// backoff strategy. Each next time tick will increase by
// a specific backoff delay.
//
// For implementation we using resource below.
// https://github.com/cenkalti/backoff/blob/v4.3.0/ticker.go.
type Ticker struct {
	// Public copy of channel which receives time ticks.
	C <-chan time.Time
	// Private channel which receives time ticks.
	c    chan time.Time
	once sync.Once
	// Channel to stop Ticker.
	stop chan struct{}

	// Duration that is added to generated backoff duration.
	baseDuration time.Duration
	// Determining the random duration of next backoff step.
	backoffSeed int

	// Timer that fires an event a certain amount of time has passed.
	timer   *time.Timer
	backoff *Backoff
}

// NewTicker creates new Ticker.
//
// You can use it as time.Ticker. The structure provides a C channel for
// receiving time tick events.
//
// Arguments:
//
//	backoff - structure for generating backoff duration.
//	minDuration - the min duration of the next tick. In most cases
//	this duration will be added to the generated backoff duration.
//	factor - a random number that will be used to generate a random
//	backoff duration. We will use numbers from 1 to 'factor' to generate
//	a random backoff based on math/random/v2 functions.
//
// Example:
//
//	func main() {
//	  // Like default time.Ticker()
//	  ticker := backoff.NewTicker(...)
//
//	  <-ticker.C
//	}
func NewTicker(backoff *Backoff, minDuration time.Duration, seed int) *Ticker {
	ch := make(chan time.Time)
	ticker := &Ticker{
		C:            ch,
		c:            ch,
		baseDuration: minDuration,
		backoffSeed:  seed,
		stop:         make(chan struct{}),
		backoff:      backoff,
	}

	go ticker.run()

	return ticker
}

func (t *Ticker) run() {
	c := t.c
	defer close(c)

	// Initialize tick countdown from time.Now().
	tickCh := t.next(time.Now())

	for {
		if tickCh == nil {
			return
		}

		select {
		case tick := <-tickCh:
			tickCh = t.next(tick)
		case <-t.stop:
			t.c = nil

			return
		}
	}
}

// next creates channel that fires after calculated
// backoff duration.
func (t *Ticker) next(tick time.Time) <-chan time.Time {
	select {
	case t.c <- tick:
	case <-t.stop:
		return nil
	}

	// Chose random seed to next backoff duration.
	backoffFactor := rand.N(backoffRange)%t.backoffSeed + 1

	next := t.backoffDuration(t.baseDuration, backoffFactor)

	return t.startTimer(next)
}

func (t *Ticker) backoffDuration(baseDuration time.Duration, backoffFactor int) time.Duration {
	next := t.backoff.Next(backoffFactor)

	return baseDuration + next
}

func (t *Ticker) startTimer(nextTick time.Duration) <-chan time.Time {
	t.resetTimer()

	t.timer = time.NewTimer(nextTick)

	return t.timer.C
}

func (t *Ticker) resetTimer() {
	if t.timer == nil {
		return
	}

	t.timer.Stop()
}

func (t *Ticker) Stop() {
	t.once.Do(t.stopTicker)
}

func (t *Ticker) stopTicker() { close(t.stop) }
