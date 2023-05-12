package inbox_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Melenium2/go-iobox/inbox"
	"github.com/Melenium2/go-iobox/inbox/mocks"
)

func TestEventMap_HandlerKeyExists(t *testing.T) {
	t.Run("should find handler with provided event type and handler key", func(t *testing.T) {
		handler := mocks.NewHandler(t)
		handler.On("Key").Return("1")

		subjects := map[string][]inbox.Handler{
			"1": {handler},
		}

		eventMap := inbox.NewEventMap(subjects)

		ok := eventMap.HandlerKeyExists("1", "1")
		assert.True(t, ok)
	})

	t.Run("should not find handler with provided event type", func(t *testing.T) {
		handler := mocks.NewHandler(t)

		subjects := map[string][]inbox.Handler{
			"2": {handler},
		}

		eventMap := inbox.NewEventMap(subjects)

		ok := eventMap.HandlerKeyExists("1", "1")
		assert.False(t, ok)
	})

	t.Run("should not find handler with provided handler key", func(t *testing.T) {
		handler := mocks.NewHandler(t)
		handler.On("Key").Return("1")

		subjects := map[string][]inbox.Handler{
			"1": {handler},
		}

		eventMap := inbox.NewEventMap(subjects)

		ok := eventMap.HandlerKeyExists("1", "2")
		assert.False(t, ok)
	})
}
