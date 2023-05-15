package inbox

import "github.com/google/uuid"

type Storage = defaultStorage

func NewStorage(conn SQLConn) *Storage {
	return newStorage(conn)
}

func NewClient(storage *Storage, handlers map[string][]Handler) Client {
	return newClient(storage, handlers)
}

var (
	id1 = uuid.New()
	id2 = uuid.New()
	id3 = uuid.New()
)

func ID1() uuid.UUID {
	return id1
}

func ID2() uuid.UUID {
	return id2
}

func ID3() uuid.UUID {
	return id3
}

func Record1() *Record {
	return &Record{
		id:         id1,
		eventType:  "1",
		handlerKey: "1",
		status:     Progress,
		payload:    []byte("{}"),
	}
}

func Record1Values() []any {
	return []any{
		Record1().id,
		Record1().eventType,
		Record1().handlerKey,
		Record1().payload,
	}
}

func Record2() *Record {
	return &Record{
		id:         id2,
		eventType:  "1",
		handlerKey: "2",
		status:     Progress,
		payload:    []byte("{}"),
	}
}

func Record2Values() []any {
	return []any{
		Record2().id,
		Record2().eventType,
		Record2().handlerKey,
		Record2().payload,
	}
}

func Record3() *Record {
	return &Record{
		id:         id3,
		eventType:  "2",
		handlerKey: "1",
		payload:    []byte("{}"),
	}
}

func Record3Values() []any {
	return []any{
		Record3().id,
		Record3().eventType,
		Record3().handlerKey,
		Record3().payload,
	}
}

type EventMap = eventMap

func NewEventMap(subjects map[string][]Handler) *EventMap {
	eventMap := newEventMap()
	eventMap.subjects = subjects

	return eventMap
}

func (r *Record) WithHandlerKey(key string) *Record {
	return r.withHandkerKey(key)
}
