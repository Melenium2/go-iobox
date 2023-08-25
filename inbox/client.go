package inbox

import "context"

// Client provides possibility to set records to the inbox table.
// All records will be processed in the future.
type Client interface {
	WriteInbox(context.Context, *Record) error
}

type client struct {
	storage  *defaultStorage
	handlers map[string][]string
}

func newClient(storage *defaultStorage, handlers map[string][]Handler) *client {
	handlerKeys := make(map[string][]string, len(handlers))

	for eventType, handlerList := range handlers {
		keys := make([]string, 0, len(handlerList))

		for _, handler := range handlerList {
			keys = append(keys, handler.Key())
		}

		handlerKeys[eventType] = keys
	}

	return &client{
		storage:  storage,
		handlers: handlerKeys,
	}
}

func (c *client) WriteInbox(ctx context.Context, record *Record) error {
	keys := c.handlers[record.eventType]

	records := make([]*Record, 0, len(keys))

	for _, key := range keys {
		records = append(records, record.withHandlerKey(key))
	}

	for _, curr := range records {
		if err := c.storage.Insert(ctx, curr); err != nil {
			return err
		}
	}

	return nil
}
