package inbox_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Melenium2/go-iobox/inbox"
	"github.com/Melenium2/go-iobox/inbox/mocks"
)

func TestRegistry_On(t *testing.T) {
	t.Run("should set new single handler", func(t *testing.T) {
		handler := mocks.NewHandler(t)
		handler.On("Key").Return("1")

		registry := inbox.NewRegistry()

		registry.On("1", handler)

		subjects := registry.Handlers()
		assert.Contains(t, subjects, "1")
	})

	t.Run("should set new multiply handlers", func(t *testing.T) {
		handler1 := mocks.NewHandler(t)
		handler1.On("Key").Return("1")

		handler2 := mocks.NewHandler(t)
		handler2.On("Key").Return("2")

		registry := inbox.NewRegistry()

		registry.On("1", handler1, handler2)

		subjects := registry.Handlers()
		assert.Contains(t, subjects, "1")

		handlers := subjects["1"]
		assert.Len(t, handlers, 2)
	})

	t.Run("should ingore handler with empty key", func(t *testing.T) {
		handler1 := mocks.NewHandler(t)
		handler1.On("Key").Return("1")

		handler2 := mocks.NewHandler(t)
		handler2.On("Key").Return("")

		registry := inbox.NewRegistry()

		registry.On("1", handler1, handler2)

		subjects := registry.Handlers()
		assert.Contains(t, subjects, "1")

		handlers := subjects["1"]
		assert.Len(t, handlers, 1)
	})

	t.Run("should ignore handler with already existed key", func(t *testing.T) {
		handler1 := mocks.NewHandler(t)
		handler1.On("Key").Return("1")

		handler2 := mocks.NewHandler(t)
		handler2.On("Key").Return("1")

		registry := inbox.NewRegistry()

		registry.On("1", handler1, handler2)

		subjects := registry.Handlers()
		assert.Contains(t, subjects, "1")

		handlers := subjects["1"]
		assert.Len(t, handlers, 1)
	})
}
