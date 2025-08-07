package inbox

import (
	"sort"
	"time"

	"github.com/google/uuid"
)

type dtoRecord struct {
	ID         string    `db:"id"`
	Status     string    `db:"status"`
	EventType  string    `db:"event_type"`
	HandlerKey string    `db:"handler_key"`
	Payload    []byte    `db:"payload"`
	Attempt    int       `db:"attempt"`
	CreatedAt  time.Time `db:"created_at"`
}

func newDtoRecord(
	id, status, eventType, handlerKey string, payload []byte, attempt int, createdAt time.Time,
) *dtoRecord {
	return &dtoRecord{
		ID:         id,
		Status:     status,
		EventType:  eventType,
		HandlerKey: handlerKey,
		Payload:    payload,
		Attempt:    attempt,
		CreatedAt:  createdAt.UTC(),
	}
}

func makeRecord(dto *dtoRecord) (*Record, error) {
	id, err := uuid.Parse(dto.ID)
	if err != nil {
		return nil, err
	}

	return newFullRecord(
		id,
		Status(dto.Status),
		dto.EventType,
		dto.HandlerKey,
		dto.Payload,
		dto.Attempt,
		dto.CreatedAt,
	), nil
}

func makeRecords(dtos []*dtoRecord) ([]*Record, error) {
	sort.Slice(dtos, func(i, j int) bool {
		t1 := dtos[i].CreatedAt
		t2 := dtos[j].CreatedAt

		return t1.Before(t2)
	})

	result := make([]*Record, len(dtos))

	for i, curr := range dtos {
		rec, err := makeRecord(curr)
		if err != nil {
			return nil, err
		}

		result[i] = rec
	}

	return result, nil
}
