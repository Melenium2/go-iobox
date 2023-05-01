package outbox

import (
	"github.com/google/uuid"
)

type dtoRecord struct {
	ID        string `db:"id"`
	Status    string `db:"status"`
	EventType string `db:"event_type"`
	Payload   []byte `db:"payload"`
}

func newDtoRecord(id, status, eventType string, payload []byte) *dtoRecord {
	return &dtoRecord{
		ID:        id,
		Status:    status,
		EventType: eventType,
		Payload:   payload,
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
