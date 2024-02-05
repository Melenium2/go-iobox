package outbox

import (
	"context"
	"database/sql"
)

type Execer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

// Client provides possibility to set outbox record to the outbox table.
// Insertion must be in the same transaction as the produced action.
type Client interface {
	WriteOutbox(context.Context, Execer, *Record) error
}

type client struct {
	storage *storage
}

func newClient(storage *storage) *client {
	return &client{
		storage: storage,
	}
}

func (c *client) WriteOutbox(ctx context.Context, tx Execer, record *Record) error {
	return c.storage.Insert(ctx, tx, record)
}
