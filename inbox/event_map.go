package inbox

import "sync"

type eventMap struct {
	mutex sync.RWMutex
	// key -> event_type, value -> handlers.
	subjects map[string][]Handler
}

func newEventMap() *eventMap {
	return &eventMap{
		subjects: make(map[string][]Handler),
	}
}

func (m *eventMap) Set(event string, handlers ...Handler) {
	m.mutex.Lock()

	m.subjects[event] = append(m.subjects[event], handlers...)

	m.mutex.Unlock()
}

func (m *eventMap) HandlerKeyExists(event, handlerKey string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	handlers, ok := m.subjects[event]
	if !ok {
		return false
	}

	for _, handler := range handlers {
		if handler.Key() == handlerKey {
			return true
		}
	}

	return false
}

func (m *eventMap) Copy() map[string][]Handler {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string][]Handler, len(m.subjects))

	for k, v := range m.subjects {
		handlers := make([]Handler, len(v))
		copy(handlers, v)

		result[k] = handlers
	}

	return result
}
