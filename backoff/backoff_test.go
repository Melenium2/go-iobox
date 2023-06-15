package backoff_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Melenium2/go-iobox/backoff"
)

func TestBackoff_Next_Should_return_duration_equals_to_min_if_first_attempt(t *testing.T) {
	b := backoff.NewBackoff()

	dur := b.Next(0)
	assert.Equal(t, backoff.DefaultMinDuration, dur)
}

func TestBackoff_Next_Should_return_max_duration_if_attempt_is_incorrect(t *testing.T) {
	b := backoff.NewBackoff()

	dur := b.Next(-1)
	assert.Equal(t, backoff.DefaultMaxDuration, dur)
}

func TestBackoff_Next_should_return_duration_more_then_provided_min_duration(t *testing.T) {
	b := backoff.NewBackoff()

	dur := b.Next(1)
	assert.Greater(t, dur, backoff.DefaultMinDuration)
}

func TestBackoff_Next_Should_return_duration_more_then_prev_attempt(t *testing.T) {
	b := backoff.NewBackoff()

	dur := b.Next(1)
	assert.Greater(t, dur, backoff.DefaultMinDuration)

	secondDur := b.Next(3)
	assert.Greater(t, secondDur, dur)
}

func TestBackoff_Next_Should_return_max_duration_from_config_if_attempt_to_high(t *testing.T) {
	b := backoff.NewBackoff()

	dur := b.Next(1e3)
	assert.Equal(t, dur, backoff.DefaultMaxDuration)
}
