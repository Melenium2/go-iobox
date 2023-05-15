package outbox

import (
	"context"
)

type Client interface {
	WriteOutbox(context.Context, SQLConn, *Record) error
}

type client struct {
	storage *defaultStorage
}

func newClient(storage *defaultStorage) *client {
	return &client{
		storage: storage,
	}
}

func (c *client) WriteOutbox(ctx context.Context, tx SQLConn, record *Record) error {
	return c.storage.Insert(ctx, tx, record)
}
