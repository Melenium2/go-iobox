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

// Registry contains all handler that will be processed by Inbox.
type Registry struct {
	eventMap *eventMap
}

func NewRegistry() *Registry {
	return &Registry{
		eventMap: newEventMap(),
	}
}

// On register new handlers to specific event key. All handlers
// will be executed on received event with provided key.
//
// Example:
//
//	  We have
//	   - event type = "order_events"
//	   - handler key = "process_order"
//	   - handler key = "update_order"
//	  The registry will bind the keys "process_order", "update_order"
//	  with event "order_events" and execute both registered handlers
//	  for each received event with event type "order_events".
//
//	  We have
//	  - event type = "order_events"
//	  - handler key = "process_order"
//	  - handler key = "process_order"
//		 If we are trying to provide several multiple handlers with the same key,
//	  then only the first handler will be associated with the event type.
//	  The second handler will be ignored.
//
//	  We have
//	  - event type = "order_events"
//	  - registered handler with key = "process_order"
//	  - new handler key = "process_order"
//	  If you are trying to provide a handler to an already existing event type, for example,
//	  "order_events", and the handler has the same key as already provided, then
//	  this handler will be ignored.
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

// Handlers returns map where key is event type and values are handlers
// associated to this event type.
func (r *Registry) Handlers() map[string][]Handler {
	return r.eventMap.Copy()
}
