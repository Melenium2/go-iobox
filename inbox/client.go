package inbox

import "context"

type Client struct {
	storage  *defaultStorage
	handlers map[string][]string
}

func newClient(storage *defaultStorage, handlers map[string][]Handler) *Client {
	handlerKeys := make(map[string][]string, len(handlers))

	for eventType, handlerList := range handlers {
		keys := make([]string, 0, len(handlerList))

		for _, handler := range handlerList {
			keys = append(keys, handler.Key())
		}

		handlerKeys[eventType] = keys
	}

	return &Client{
		storage:  storage,
		handlers: handlerKeys,
	}
}

func (c *Client) WriteInbox(ctx context.Context, record *Record) error {
	keys := c.handlers[record.eventType]

	records := make([]*Record, 0, len(keys))

	for _, key := range keys {
		records = append(records, record.withHandkerKey(key))
	}

	for _, curr := range records {
		if err := c.storage.Insert(ctx, curr); err != nil {
			return err
		}
	}

	return nil
}
