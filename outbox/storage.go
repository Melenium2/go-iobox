package outbox

import (
	"context"
	"database/sql"
	"strings"
)

type defaultStorage struct {
	conn SQLConn
}

func newStorage(conn SQLConn) *defaultStorage {
	return &defaultStorage{
		conn: conn,
	}
}

func (s *defaultStorage) InitOutboxTable(ctx context.Context) error {
	sql := "create table if not exists __outbox_table " +
		" 	(" +
		" 		id varchar(36) not null primary key," +
		" 		status varchar(12)," +
		" 		event_type varchar(255) not null," +
		" 		payload jsonb not null," +
		" 		created_at timestamp not null default (now() at time zone 'utc')," +
		" 		updated_at timestamp not null default (now() at time zone 'utc')" +
		" 	);"

	_, err := s.conn.ExecContext(ctx, sql)

	return err
}

func (s *defaultStorage) Fetch(ctx context.Context) ([]*Record, error) {
	dest := make([]*dtoRecord, 0)

	sqlStr := "update __outbox_table set " +
		" 				status = $1," +
		" 				updated_at = (now() at time zone 'utc') " +
		" 		where status is null " +
		" 		returning id, status, event_type, payload;"

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
			" 		where id in ($2);"
		recordsStatus sql.NullString
	)

	if records[0].status != "" {
		recordsStatus = sql.NullString{String: string(records[0].status), Valid: true}
	}

	ids := make([]string, len(records))

	for i := 0; i < len(records); i++ {
		ids[i] = records[i].id.String()
	}

	_, err := s.conn.ExecContext(ctx, sqlStr, recordsStatus, strings.Join(ids, ", "))

	return err
}

func (s *defaultStorage) Insert(ctx context.Context, tx SQLConn, record *Record) error {
	sqlStr := "insert into __outbox_table (id, event_type, payload) values ($1, $2, $3);"

	payload, err := record.payload.MarshalJSON()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, sqlStr, record.id, record.eventType, string(payload))

	return err
}

func (s *defaultStorage) selectRows(ctx context.Context, conn SQLConn, dest *[]*dtoRecord, sqlStr string, args ...any) error {
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
	)
	for rows.Next() {
		err = rows.Scan(&id, &status, &eventType, &payload)
		if err != nil {
			return err
		}

		*dest = append(*dest, newDtoRecord(id, status.String, eventType, payload))
	}

	return nil
}
