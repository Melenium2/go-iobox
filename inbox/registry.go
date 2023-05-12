package inbox

import "context"

//go:generate mockery --name Handler
type Handler interface {
	Key() string
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
