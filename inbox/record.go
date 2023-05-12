package inbox

import (
	"fmt"

	"github.com/google/uuid"
)

type Status string

const (
	Progress Status = "progress"
	Failed   Status = "failed"
	Done     Status = "done"
	Null     Status = ""
)

type Record struct {
	id         uuid.UUID
	eventType  string
	handlerKey string
	status     Status
	payload    []byte
}

func NewRecord(id uuid.UUID, eventType string, payload []byte) (*Record, error) {
	if eventType == "" {
		return nil, fmt.Errorf("incorrect record event type provided")
	}

	return &Record{
		id:        id,
		eventType: eventType,
		payload:   payload,
	}, nil
}

func newFullRecord(id uuid.UUID, status Status, eventType, handlerKey string, payload []byte) *Record {
	return &Record{
		id:         id,
		status:     status,
		eventType:  eventType,
		handlerKey: handlerKey,
		payload:    payload,
	}
}

func (r *Record) Done() {
	r.status = Done
}

func (r *Record) Fail() {
	r.status = Failed
}

func (r *Record) Null() {
	r.status = ""
}

func (r *Record) withHandkerKey(key string) *Record {
	b := make([]byte, len(r.payload))
	copy(b, r.payload)

	return &Record{
		id:         r.id,
		eventType:  r.eventType,
		handlerKey: key,
		status:     r.status,
		payload:    b,
	}
}
