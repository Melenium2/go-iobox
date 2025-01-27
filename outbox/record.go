package outbox

import (
	"encoding/json"
)

// Status defines current status of Record.
type Status string

const (
	// Progress means the current Record is processed by outbox worker.
	Progress Status = "progress"
	// Failed means the current Record not processed by worker by specific
	// reason.
	Failed Status = "failed"
	// Done means the current Record is successfully processed.
	Done Status = "done"
	// Null means the current Record is not processed yet.
	Null Status = ""
)

// Record is event that should be processed by outbox worker.
type Record struct {
	id        string
	eventType string
	status    Status
	payload   json.Marshaler
}

// NewRecord creates new record that can be processed by outbox worker.
//
// Parameters:
//
//	id - is a unique id for outbox table. ID should be unique or storage
//			will ignore all duplicate ids. ID can container max 36 byte.
//	eventType - is a topic to which event will be published.
//	payload - the body to be published.
func NewRecord(id string, eventType string, payload json.Marshaler) *Record {
	return &Record{
		id:        id,
		eventType: eventType,
		payload:   payload,
	}
}

func newFullRecord(
	id string,
	status Status,
	eventType string,
	payload json.Marshaler,
) *Record {
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

// Null sets Null status to current Record.
func (r *Record) Null() {
	r.status = ""
}
