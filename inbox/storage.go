package inbox

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Melenium2/go-iobox/inbox/migrations"
	"github.com/Melenium2/go-iobox/migration"
)

const tableName = "__inbox_table"

type defaultStorage struct {
	conn *sql.DB
}

func newStorage(conn *sql.DB) *defaultStorage {
	return &defaultStorage{
		conn: conn,
	}
}

func (s *defaultStorage) InitInboxTable(ctx context.Context) error {
	m := migration.New()

	if err := m.SetupFS(ctx, s.conn, migrations.FS, "inbox_schema"); err != nil {
		return fmt.Errorf("failed to setup inbox migrations, %w", err)
	}

	err := m.Up()
	if err == nil {
		return nil
	}

	_ = m.Down()

	return fmt.Errorf("failed to run migrations, %w", err)
}

func (s *defaultStorage) Fetch(ctx context.Context, fetchTime time.Time) ([]*Record, error) {
	dest := make([]*dtoRecord, 0)

	sqlStr := "update " + tableName + " set " +
		" 				status = $1," +
		" 				updated_at = (now() at time zone 'utc') " +
		" 		where " +
		" 			status is null or " +
		" 			(status = 'failed' and next_attempt <= $2) " +
		" 		returning id, status, event_type, handler_key, payload, attempt, created_at;"

	if err := s.selectRows(ctx, s.conn, &dest, sqlStr, Progress, fetchTime); err != nil {
		return nil, fmt.Errorf("error while fetching records, %w", err)
	}

	if len(dest) == 0 {
		return nil, ErrNoRecords
	}

	return makeRecords(dest)
}

func (s *defaultStorage) Update(ctx context.Context, records []*Record) error {
	if len(records) == 0 {
		return nil
	}

	sqlStr := "update " + tableName + " set " +
		" 			status = $1, " +
		" 			attempt = $2, " +
		" 			error_message = $3, " +
		" 			next_attempt = $4, " +
		"			updated_at = (now() at time zone 'utc') " +
		" 		where id = $5 and handler_key = $6;"

	for i := range len(records) {
		var (
			recordStatus    sql.NullString
			errorMessage    sql.NullString
			attemptDeadline sql.NullTime
		)

		curr := records[i]

		if curr.status != "" {
			recordStatus = sql.NullString{String: string(curr.status), Valid: true}
		}

		if curr.attempt.message != "" {
			errorMessage = sql.NullString{String: curr.attempt.message, Valid: true}
		}

		if !curr.attempt.nextAttempt.IsZero() {
			attemptDeadline = sql.NullTime{Time: curr.attempt.nextAttempt, Valid: true}
		}

		_, err := s.conn.ExecContext(
			ctx,
			sqlStr,
			recordStatus,
			curr.attempt.attempt,
			errorMessage,
			attemptDeadline,
			curr.id,
			curr.handlerKey,
		)
		if err != nil {
			return fmt.Errorf("error while updating records, %w", err)
		}
	}

	return nil
}

func (s *defaultStorage) Insert(ctx context.Context, record *Record) error {
	sqlStr := "insert into " + tableName + " (id, event_type, handler_key, payload, created_at) " +
		" values ($1, $2, $3, $4, $5) on conflict (id, handler_key) do nothing;"

	_, err := s.conn.ExecContext(
		ctx,
		sqlStr,
		record.id,
		record.eventType,
		record.handlerKey,
		record.payload,
		record.eventDate,
	)

	return err
}

func (s *defaultStorage) selectRows(
	ctx context.Context, conn *sql.DB, dest *[]*dtoRecord, sqlStr string, args ...any,
) error {
	rows, err := conn.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return err
	}

	if err := rows.Err(); err != nil {
		return err
	}

	defer rows.Close()

	var (
		id         string
		status     sql.NullString
		eventType  string
		handlerKey string
		payload    []byte
		attempt    int
		createdAt  time.Time
	)

	for rows.Next() {
		err = rows.Scan(&id, &status, &eventType, &handlerKey, &payload, &attempt, &createdAt)
		if err != nil {
			return err
		}

		dto := newDtoRecord(id, status.String, eventType, handlerKey, payload, attempt, createdAt)

		*dest = append(*dest, dto)
	}

	return nil
}
