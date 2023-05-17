package inbox

import (
	"fmt"

	"github.com/google/uuid"
)

// Status defines current status of Record.
type Status string

const (
	// Progress means the current Record is processed by worker.
	Progress Status = "progress"
	// Failed means the current Record not processed by worker by specific
	// reason.
	Failed Status = "failed"
	// Done means the current Record is successfully processed.
	Done Status = "done"
	// Null means the current Record is not processed yet.
	Null Status = ""
)

// Record is event that should be processed by inbox worker.
type Record struct {
	id         uuid.UUID
	eventType  string
	handlerKey string
	status     Status
	payload    []byte
}

// NewRecord creates new record that can be processed by inbox worker.
//
// Parameters:
//
//	id - is a unique id for inbox table. ID should be unique or storage
//			will ignore all duplicate ids.
//	eventType - is a topic with which event was published.
//	payload - the received body.
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

// Done sets Done status to current Record. Status will be
// ignored on first save to the outbox table.
func (r *Record) Done() {
	r.status = Done
}

// Fail sets Failed status to current Record. Status will be
// ignored on first save to the outbox table.
func (r *Record) Fail() {
	r.status = Failed
}

// Null sets Null status to current Record.
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
