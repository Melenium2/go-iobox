package inbox

import "context"

//go:generate mockery --name Handler
type Handler interface {
	// Key is a unique identifier of current handler.
	// This string must be not empty and must be unique for each
	// handler that passed to the Registry. Only the first handler with a key
	// will be stored in the Registry, all other handlers with the same key
	// will be ignored.
	Key() string
	// Process is a function that will be executed for each handler associated
	// with specific event_type and key provided by the Handler implementation.
	Process(context.Context, []byte) error
}

type Registry struct {
	eventMap *eventMap
}

func NewRegistry() *Registry {
	return &Registry{
		eventMap: newEventMap(),
	}
}

func (r *Registry) On(event string, handlers ...Handler) {
	var (
		correctHandlers = make([]Handler, 0)
		existed         = make(map[string]struct{}, 0)
	)

	for _, handler := range handlers {
		if handler.Key() == "" {
			continue
		}

		if _, ok := existed[handler.Key()]; ok {
			continue
		}

		if r.eventMap.HandlerKeyExists(event, handler.Key()) {
			continue
		}

		correctHandlers = append(correctHandlers, handler)
		existed[handler.Key()] = struct{}{}
	}

	r.eventMap.Set(event, correctHandlers...)
}

func (r *Registry) Handlers() map[string][]Handler {
	return r.eventMap.Copy()
}
