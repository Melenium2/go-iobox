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

type Logger interface {
	Print(...any)
	Printf(string, ...any)
}

// Outbox is struct that implement of outbox pattern.
//
// Writing all outgoing events in a temporary table in the same transaction
// in which we process the action associated with this event.
// Then we try to publish the event in the broker with specific timeout
// until the event is sent.
//
// More about outbox pattern you can read at
// https://microservices.io/patterns/data/transactional-outbox.html.
type Outbox struct {
	broker Broker
	logger Logger

	storage *defaultStorage
	config  config
}

// NewOutbox creates new outbox implementation.
func NewOutbox(broker Broker, conn SQLConn, opts ...Option) *Outbox {
	defaultCfg := defaultConfig()

	for _, opt := range opts {
		defaultCfg = opt(defaultCfg)
	}

	return &Outbox{
		broker:  broker,
		logger:  defaultCfg.logger,
		storage: newStorage(conn),
		config:  defaultCfg,
	}
}

// Writer creates new Client to write outgoing events to the temporary table.
func (o *Outbox) Writer() Client {
	return newClient(o.storage)
}

// Start initialize outbox table and start worker process. Worker
// is process that send outgoing messages to broker.
//
// Start function blocks current thread.
func (o *Outbox) Start(ctx context.Context) error {
	if err := o.storage.InitOutboxTable(ctx); err != nil {
		return fmt.Errorf("can not initialize outbox table, stroage return err: %w", err)
	}

	go o.run()

	return nil
}

// run starts the publishing process.
func (o *Outbox) run() {
	ticker := time.NewTicker(o.config.iterationRate)

	for range ticker.C {
		if err := o.iteration(context.Background()); err != nil {
			o.logger.Print(err.Error())
		}
	}
}

// iteration tries to send events to the broker, if operation was successful
// updates status in the outbox table.
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
		o.logger.Printf(
			"count of records does not match, len %d, success %d, fail %d",
			len(records), len(success), len(fail),
		)
	}

	if err := o.storage.Update(ctx, success); err != nil {
		return err
	}

	if err := o.storage.Update(ctx, fail); err != nil {
		return err
	}

	return nil
}
