package outbox

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type PayloadMarshaler = dtoPayload

type Storage = defaultStorage

var (
	id1 = uuid.New()
	id2 = uuid.New()
	id3 = uuid.New()
)

func ID1() uuid.UUID {
	return id1
}

func ID2() uuid.UUID {
	return id2
}

func ID3() uuid.UUID {
	return id3
}

func Record1() *Record {
	payload := dtoPayload{Body: []byte("{}")}

	return newFullRecord(id1, Progress, "topic1", &payload)
}

func Record2() *Record {
	payload := dtoPayload{Body: []byte("{}")}

	return newFullRecord(id2, Progress, "topic1", &payload)
}

func Record3() *Record {
	payload := dtoPayload{Body: []byte("{}")}

	return newFullRecord(id3, Done, "topic1", &payload)
}

func NewStorage(conn *sql.DB) *defaultStorage {
	return newStorage(conn)
}

func (o *Outbox) Iteration() error {
	return o.iteration(context.Background())
}
