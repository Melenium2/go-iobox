package outbox

import (
	"encoding/json"

	"github.com/google/uuid"
)

// Status defines current status of Record.
type Status string

const (
	// Progress means the current Record is process by outbox worker.
	Progress Status = "progress"
	// Failed means the current Record not processed by worker by specific
	// reason.
	Failed Status = "failed"
	// Done means the current Record is successfully processed.
	Done Status = "done"
)

// Record is event that should be processed by outbox worker.
type Record struct {
	id        uuid.UUID
	eventType string
	status    Status
	payload   json.Marshaler
}

// NewRecord creates new record that can be processed by outbox worker.
//
// Parameters:
//
//	id - is a unique id for outbox table. ID should be unique or storage
//			will ignore all duplicate ids.
//	eventType - is a topic to which event will be published.
//	payload - the body to be published.
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
