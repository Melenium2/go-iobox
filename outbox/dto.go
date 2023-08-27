package outbox

import (
	"sort"
	"time"

	"github.com/google/uuid"
)

type dtoRecord struct {
	ID        string    `db:"id"`
	Status    string    `db:"status"`
	EventType string    `db:"event_type"`
	Payload   []byte    `db:"payload"`
	CreatedAt time.Time `db:"created_at"`
}

func newDtoRecord(id, status, eventType string, payload []byte, createdAt time.Time) *dtoRecord {
	return &dtoRecord{
		ID:        id,
		Status:    status,
		EventType: eventType,
		Payload:   payload,
		CreatedAt: createdAt,
	}
}

type dtoPayload struct {
	Body []byte
}

func (dp *dtoPayload) MarshalJSON() ([]byte, error) {
	return dp.Body, nil
}

func makeRecord(dto *dtoRecord) (*Record, error) {
	id, err := uuid.Parse(dto.ID)
	if err != nil {
		return nil, err
	}

	payload := dtoPayload{Body: dto.Payload}

	return newFullRecord(id, Status(dto.Status), dto.EventType, &payload), nil
}

func makeRecords(dtos []*dtoRecord) ([]*Record, error) {
	sort.Slice(dtos, func(i, j int) bool {
		t1 := dtos[i].CreatedAt
		t2 := dtos[j].CreatedAt

		return t1.Before(t2)
	})

	result := make([]*Record, len(dtos))

	for i := 0; i < len(dtos); i++ {
		rec, err := makeRecord(dtos[i])
		if err != nil {
			return nil, err
		}

		result[i] = rec
	}

	return result, nil
}
