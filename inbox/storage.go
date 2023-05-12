package inbox

import (
	"context"
	"database/sql"
)

type defaultStorage struct {
	conn SQLConn
}

func newStorage(conn SQLConn) *defaultStorage {
	return &defaultStorage{
		conn: conn,
	}
}

func (s *defaultStorage) InitInboxTable(ctx context.Context) error {
	sql := "create table if not exists __inbox_table " +
		" 	(" +
		" 		id varchar(36) not null," +
		" 		status varchar(12)," +
		" 		event_type varchar(255) not null," +
		" 		handler_key varchar(255) not null, " +
		" 		payload bytea not null default '{}'::bytea," +
		" 		created_at timestamp not null default (now() at time zone 'utc')," +
		" 		updated_at timestamp not null default (now() at time zone 'utc')" +
		" 	);" +
		"" +
		"create unique index if not exists __inbox_uniq_id_handler_key_idx on __inbox_table (id, handler_key);"

	_, err := s.conn.ExecContext(ctx, sql)

	return err
}

func (s *defaultStorage) Fetch(ctx context.Context) ([]*Record, error) {
	dest := make([]*dtoRecord, 0)

	sqlStr := "update __inbox_table set " +
		" 				status = $1," +
		" 				updated_at = (now() at time zone 'utc') " +
		" 		where status is null " +
		" 		returning id, status, event_type, handler_key, payload;"

	if err := s.selectRows(ctx, s.conn, &dest, sqlStr, Progress); err != nil {
		return nil, err
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

	sqlStr := "update __inbox_table set " +
		" 			status = $1, " +
		"			updated_at = (now() at time zone 'utc') " +
		" 		where id = $2 and handler_key = $3;"

	stmt, err := s.conn.PrepareContext(ctx, sqlStr)
	if err != nil {
		return err
	}

	defer stmt.Close()

	for i := 0; i < len(records); i++ {
		var recordStatus sql.NullString

		curr := records[i]

		if curr.status != "" {
			recordStatus = sql.NullString{String: string(curr.status), Valid: true}
		}

		_, err = stmt.ExecContext(ctx, recordStatus, curr.id, curr.handlerKey)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *defaultStorage) Insert(ctx context.Context, record *Record) error {
	sqlStr := "insert into __inbox_table (id, event_type, handler_key, payload) " +
		" values ($1, $2, $3, $4) on conflict (id, handler_key) do nothing;"

	_, err := s.conn.ExecContext(
		ctx, sqlStr, record.id, record.eventType, record.handlerKey, record.payload,
	)

	return err
}

func (s *defaultStorage) selectRows(
	ctx context.Context, conn SQLConn, dest *[]*dtoRecord, sqlStr string, args ...any,
) error {
	rows, err := conn.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return err
	}

	defer rows.Close()

	var (
		id         string
		status     sql.NullString
		eventType  string
		handlerKey string
		payload    []byte
	)
	for rows.Next() {
		err = rows.Scan(&id, &status, &eventType, &handlerKey, &payload)
		if err != nil {
			return err
		}

		*dest = append(*dest, newDtoRecord(id, status.String, eventType, handlerKey, payload))
	}

	return nil
}
