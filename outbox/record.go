package outbox

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Status string

const (
	Progress Status = "progress"
	Failed   Status = "failed"
	Done     Status = "done"
)

type Record struct {
	id        uuid.UUID
	eventType string
	status    Status
	payload   json.Marshaler
}

func NewRecord(id uuid.UUID, eventType string, payload json.Marshaler) *Record {
	return &Record{
		id:        id,
		eventType: eventType,
		payload:   payload,
	}
}

func newFullRecord(id uuid.UUID, status Status, eventType string, payload json.Marshaler) *Record {
	return &Record{
		id:        id,
		status:    status,
		eventType: eventType,
		payload:   payload,
	}
}

func (r *Record) Done() {
	r.status = Done
}

func (r *Record) Fail() {
	r.status = Failed
}
