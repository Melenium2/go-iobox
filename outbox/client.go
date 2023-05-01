package outbox

import (
	"context"
)

type Inserter interface {
	Insert(ctx context.Context, tx SQLConn, record *Record) error
}

type Client struct {
	storage Inserter
}

func NewClient(storage Inserter) *Client {
	return &Client{
		storage: storage,
	}
}

func (c *Client) WriteRecord(ctx context.Context, tx SQLConn, record *Record) error {
	return c.storage.Insert(ctx, tx, record)
}
