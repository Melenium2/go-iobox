package inbox

import (
	"context"
	"database/sql"
	"time"
)

//go:generate mockery --name SQLConn
type SQLConn interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
}

type Logger interface {
	Print(...any)
	Printf(string, ...any)
}

type Inbox struct {
	handlers map[string][]Handler
	storage  *defaultStorage
	logger   Logger
	config   config
}

func NewInbox(registry *Registry, conn SQLConn, opts ...Option) *Inbox {
	cfg := defaultConfig()

	for _, opt := range opts {
		cfg = opt(cfg)
	}

	return &Inbox{
		handlers: registry.Handlers(),
		storage:  newStorage(conn),
		logger:   cfg.logger,
		config:   cfg,
	}
}

func (i *Inbox) Writer() Client {
	return newClient(i.storage, i.handlers)
}

func (i *Inbox) Start(ctx context.Context) error {
	if err := i.storage.InitInboxTable(ctx); err != nil {
		return err
	}

	go i.run()

	return nil
}

func (i *Inbox) run() {
	ticker := time.NewTicker(i.config.iterationRate)

	for range ticker.C {
		if err := i.iteration(); err != nil {
			i.logger.Print(err.Error())
		}
	}
}

func (i *Inbox) iteration() error {
	ctx := context.Background()

	records, err := i.storage.Fetch(ctx)
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
			record.Fail()

			continue
		}

		record.Done()
	}

	return i.updateStatus(ctx, records)
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

func (i *Inbox) updateStatus(ctx context.Context, records []*Record) error {
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
		i.logger.Printf(
			"count of recrods does not match, len %d, success %d, fail %d",
			len(records), len(success), len(fail),
		)
	}

	if err := i.storage.Update(ctx, success); err != nil {
		return err
	}

	if err := i.storage.Update(ctx, fail); err != nil {
		return err
	}

	return nil
}
