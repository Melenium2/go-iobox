package inbox

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Melenium2/go-iobox/backoff"
	"github.com/Melenium2/go-iobox/retention"
)

// Inbox is struct that implement inbox pattern.
//
// Writing all incoming events in a temporary table to future processing.
// Then we try to process each event with the provided handlers.
// In addition, Inbox filters new events. All events with the same event_id
// will be ignored.
//
// More about inbox pattern you can read at
// https://softwaremill.com/microservices-101.
type Inbox struct {
	config config

	handlers  map[string][]Handler
	storage   *defaultStorage
	backoff   *backoff.Backoff
	retention *retention.Policy
}

func NewInbox(registry *Registry, conn *sql.DB, opts ...Option) *Inbox {
	cfg := defaultConfig()

	for _, opt := range opts {
		cfg = opt(cfg)
	}

	return &Inbox{
		handlers:  registry.Handlers(),
		storage:   newStorage(conn),
		config:    cfg,
		backoff:   backoff.NewBackoff(),
		retention: retention.NewPolicy(conn, tableName, cfg.retention),
	}
}

// Writer creates new Client to store incoming events to the temporary table.
func (i *Inbox) Writer() Client {
	return newClient(i.storage, i.handlers)
}

// Start creates new inbox table if it not created and starts worker
// which process records from the table. To stop inbox worker, you can
// call context close() function.
func (i *Inbox) Start(ctx context.Context) error {
	if err := i.storage.InitInboxTable(ctx); err != nil {
		return err
	}

	go i.run(ctx)
	go i.retention.Start(ctx)

	return nil
}

func (i *Inbox) run(ctx context.Context) {
	var (
		backoffConfig = backoff.Config{
			Min: time.Second,
			Max: i.config.iterationRate,
		}
		bf = backoff.NewBackoff(backoffConfig)
	)

	ticker := backoff.NewTicker(bf, i.config.iterationRate, i.config.iterationSeed)

	for {
		select {
		case <-ticker.C:
			err := i.iteration() //nolint:contextcheck
			if errors.Is(err, ErrNoRecords) {
				continue
			}

			if err != nil {
				i.config.onError(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// iteration fetches all incoming events from a temporary table
// and trying to process it. In some cases the worker can not process
// incoming events. 1) If we received an unknown event_type. 2) If the handler with
// required key not found in the Registry. In this cases we skip
// current record and sets its status to Null. In the next iteration
// we again try to handle the event. In other cases we set Fail or
// Done status to the record depends on in the result of handler.
func (i *Inbox) iteration() error {
	ctx := context.Background()

	records, err := i.storage.Fetch(ctx, time.Now().UTC())
	if err != nil {
		return err
	}

	for _, record := range records {
		handlers, ok := i.handlers[record.eventType]
		if !ok {
			record.Null()

			continue
		}

		handler, ok := i.lookForHandler(record.handlerKey, handlers)
		if !ok {
			record.Null()

			continue
		}

		if err = i.process(ctx, handler, record.payload); err != nil {
			// function mutate record inside itself.
			_ = i.failOrDead(record, err)

			continue
		}

		record.Done()
	}

	return i.storage.Update(ctx, records)
}

func (i *Inbox) lookForHandler(handlerKey string, handlers []Handler) (Handler, bool) {
	for _, handler := range handlers {
		if handler.Key() == handlerKey {
			return handler, true
		}
	}

	return nil, false
}

func (i *Inbox) process(ctx context.Context, handler Handler, payload []byte) error {
	ctx, cancel := context.WithTimeout(ctx, i.config.handlerTimeout)
	defer cancel()

	return handler.Process(ctx, payload)
}

func (i *Inbox) failOrDead(record *Record, err error) *Record {
	record.Fail(err)

	attempt := record.Attempt()

	if attempt >= i.config.maxRetryAttempts {
		record.Dead()

		i.config.onDead(record.id, err.Error())

		return record
	}

	dur := i.backoff.Next(attempt)

	record.CalcNewDeadline(dur)

	return record
}
