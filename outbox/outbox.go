package outbox

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Broker interface {
	Publish(ctx context.Context, subject string, payload []byte) error
}

type SQLConn interface {
	ExecContext(ctx context.Context, sql string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, sql string, args ...any) (*sql.Rows, error)
}

type Outbox struct {
	broker Broker

	storage *defaultStorage
	config  config
}

func NewOutbox(broker Broker, conn SQLConn, opts ...Option) *Outbox {
	defaultCfg := defaultConfig()

	for _, opt := range opts {
		defaultCfg = opt(defaultCfg)
	}

	return &Outbox{
		broker:  broker,
		storage: newStorage(conn),
		config:  defaultCfg,
	}
}

func (o *Outbox) Writer() *Client {
	return NewClient(o.storage)
}

// Start initialize outbox table and start worker process. Worker porcess
// is process that send incoming outbox messages to broker.
//
// Start function blocks current thread.
func (o *Outbox) Start(ctx context.Context) error {
	if err := o.storage.InitOutboxTable(ctx); err != nil {
		return fmt.Errorf("can not initialize outbox table, stroage return err: %w", err)
	}

	go o.run()

	return nil
}

func (o *Outbox) run() {
	ticker := time.NewTicker(o.config.iterationRate)

	for range ticker.C {
		if err := o.iteration(context.Background()); err != nil {
			// todo may be log here.
		}
	}
}

func (o *Outbox) iteration(ctx context.Context) error {
	records, err := o.storage.Fetch(ctx)
	if err != nil {
		return err
	}

	for _, record := range records {
		payload, err := record.payload.MarshalJSON()
		if err != nil {
			record.Fail()

			return err
		}

		if err := o.broker.Publish(ctx, record.eventType, payload); err != nil {
			record.Fail()

			return err
		}

		record.Done()
	}

	if err := o.updateStatus(ctx, records); err != nil {
		return err
	}

	return nil
}

func (o *Outbox) updateStatus(ctx context.Context, records []*Record) error {
	success := make([]*Record, 0)
	fail := make([]*Record, 0)

	for _, record := range records {
		if record.status == Done {
			success = append(success, record)
		}

		if record.status == Failed {
			fail = append(fail, record)
		}
	}

	if len(success)+len(fail) != len(records) {
		// todo remove this condition or maybe write log.
	}

	if err := o.storage.Update(ctx, success); err != nil {
		return err
	}

	if err := o.storage.Update(ctx, fail); err != nil {
		return err
	}

	return nil
}
