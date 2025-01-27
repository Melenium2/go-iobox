package outbox

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type PayloadMarshaler = dtoPayload

type Storage = defaultStorage

var (
	id1 = uuid.NewString()
	id2 = uuid.NewString()
	id3 = uuid.NewString()
)

func ID1() string {
	return id1
}

func ID2() string {
	return id2
}

func ID3() string {
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

type DTORecord = dtoRecord

func MakeRecrods(dtos []*DTORecord) ([]*Record, error) {
	return makeRecords(dtos)
}
