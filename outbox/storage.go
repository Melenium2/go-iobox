package outbox

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/Melenium2/go-iobox/migration"
	"github.com/Melenium2/go-iobox/outbox/migrations"
)

type defaultStorage struct {
	conn *sql.DB
}

func newStorage(conn *sql.DB) *defaultStorage {
	return &defaultStorage{
		conn: conn,
	}
}

func (s *defaultStorage) InitOutboxTable(ctx context.Context) error {
	m := migration.New()

	if err := m.SetupFS(ctx, s.conn, migrations.FS, "outbox_schema"); err != nil {
		return fmt.Errorf("failed to setup outbox migrations, %w", err)
	}

	err := m.Up()
	if err == nil {
		return nil
	}

	_ = m.Down()

	return fmt.Errorf("failed to run migrations, %w", err)
}

func (s *defaultStorage) Fetch(ctx context.Context) ([]*Record, error) {
	dest := make([]*dtoRecord, 0)

	sqlStr := "update __outbox_table set " +
		" 				status = $1," +
		" 				updated_at = (now() at time zone 'utc') " +
		" 		where status is null " +
		" 		returning id, status, event_type, payload, created_at;"

	if err := s.selectRows(ctx, s.conn, &dest, sqlStr, Progress); err != nil {
		return nil, err
	}

	if len(dest) == 0 {
		return nil, ErrNoRecrods
	}

	return makeRecords(dest)
}

func (s *defaultStorage) Update(ctx context.Context, records []*Record) error {
	if len(records) == 0 {
		return nil
	}

	var (
		sqlStr = "update __outbox_table set " +
			" 			status = $1, " +
			"			updated_at = (now() at time zone 'utc') " +
			" 		where id = any ($2);"

		recordsStatus sql.NullString
	)

	if records[0].status != "" {
		recordsStatus = sql.NullString{String: string(records[0].status), Valid: true}
	}

	ids := make([]string, len(records))

	for i := 0; i < len(records); i++ {
		ids[i] = records[i].id.String()
	}

	_, err := s.conn.ExecContext(ctx, sqlStr, recordsStatus, pq.Array(ids))

	return err
}

func (s *defaultStorage) Insert(ctx context.Context, tx Execer, record *Record) error {
	sqlStr := "insert into __outbox_table (id, event_type, payload) values ($1, $2, $3) " +
		" on conflict do nothing;"

	payload, err := record.payload.MarshalJSON()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, sqlStr, record.id, record.eventType, string(payload))

	return err
}

func (s *defaultStorage) selectRows(ctx context.Context, conn *sql.DB, dest *[]*dtoRecord, sqlStr string, args ...any) error {
	rows, err := conn.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return err
	}

	defer rows.Close()

	var (
		id        string
		status    sql.NullString
		eventType string
		payload   []byte
		createdAt time.Time
	)
	for rows.Next() {
		err = rows.Scan(&id, &status, &eventType, &payload, &createdAt)
		if err != nil {
			return err
		}

		*dest = append(*dest, newDtoRecord(id, status.String, eventType, payload, createdAt))
	}

	return nil
}
