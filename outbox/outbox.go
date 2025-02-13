package outbox

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Melenium2/go-iobox/backoff"
	"github.com/Melenium2/go-iobox/retention"
)

type Broker interface {
	Publish(ctx context.Context, subject string, payload []byte) error
}

// Outbox is struct that implement outbox pattern.
//
// Writing all outgoing events in a temporary table in the same transaction
// in which we process the action associated with this event.
// Then we try to publish the event to the broker with specific timeout
// until the event is sent.
//
// More about outbox pattern you can read at
// https://microservices.io/patterns/data/transactional-outbox.html.
type Outbox struct {
	config config

	broker    Broker
	storage   *defaultStorage
	retention *retention.Policy
}

// NewOutbox creates new outbox implementation.
func NewOutbox(broker Broker, conn *sql.DB, opts ...Option) *Outbox {
	cfg := defaultConfig()

	for _, opt := range opts {
		cfg = opt(cfg)
	}

	return &Outbox{
		broker:    broker,
		storage:   newStorage(conn),
		retention: retention.NewPolicy(conn, tableName, cfg.retention),
		config:    cfg,
	}
}

// Writer creates new Client to write outgoing events to the temporary table.
func (o *Outbox) Writer() Client {
	return newClient(o.storage)
}

// Start initialize outbox table and start worker process. Worker
// is process that send outgoing messages to broker.
func (o *Outbox) Start(ctx context.Context) error {
	if err := o.storage.InitOutboxTable(ctx); err != nil {
		return fmt.Errorf("can not initialize outbox table, storage return err: %w", err)
	}

	go o.run(ctx)
	go o.retention.Start(ctx)

	return nil
}

// run starts the publishing process.
func (o *Outbox) run(ctx context.Context) {
	var (
		backoffConfig = backoff.Config{
			Min: time.Second,
			Max: o.config.iterationRate,
		}
		bf = backoff.NewBackoff(backoffConfig)
	)

	ticker := backoff.NewTicker(bf, o.config.iterationRate, o.config.iterationSeed)

	for {
		select {
		case <-ticker.C:
			if err := o.iteration(context.Background()); err != nil {
				o.config.onError(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// iteration tries to send events to the broker, if operation was successful
// updates status in the outbox table.
func (o *Outbox) iteration(ctx context.Context) error {
	records, err := o.storage.Fetch(ctx)
	if errors.Is(err, ErrNoRecrods) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("records not fetched, %w", err)
	}

	for _, record := range records {
		record.Done()

		payload, err := record.payload.MarshalJSON()
		if err != nil {
			record.Fail()

			return fmt.Errorf("payload not marshaled, %w", err)
		}

		if err := o.publish(ctx, record.eventType, payload); err != nil {
			// If we can not publish the event during a connection issue
			// or whatever, we set the current record status to Null.
			// This means that the current record has not yet been published.
			record.Null()

			o.config.onError(err)
		}
	}

	if err := o.updateStatus(ctx, records); err != nil {
		return err
	}

	return nil
}

func (o *Outbox) publish(ctx context.Context, eventType string, payload []byte) error {
	ctx, cancel := context.WithTimeout(ctx, o.config.timeout)
	defer cancel()

	err := o.broker.Publish(ctx, eventType, payload)
	if err != nil {
		return fmt.Errorf("event %q not published, %w", eventType, err)
	}

	return nil
}

func (o *Outbox) updateStatus(ctx context.Context, records []*Record) error {
	var (
		success = make([]*Record, 0)
		fail    = make([]*Record, 0)
		null    = make([]*Record, 0)
	)

	for _, record := range records {
		if record.status == Done {
			success = append(success, record)
		}

		if record.status == Failed {
			fail = append(fail, record)
		}

		if record.status == Null {
			null = append(null, record)
		}
	}

	if err := o.storage.Update(ctx, success); err != nil {
		return err
	}

	if err := o.storage.Update(ctx, fail); err != nil {
		return err
	}

	if err := o.storage.Update(ctx, null); err != nil {
		return err
	}

	return nil
}
