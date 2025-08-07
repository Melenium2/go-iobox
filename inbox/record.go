package inbox

import (
	"fmt"
	"time"

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
	// Dead means the current Record is not processable.
	Dead Status = "dead"
)

type attempt struct {
	attempt     int
	message     string
	nextAttempt time.Time
}

// Record is event that should be processed by inbox worker.
type Record struct {
	id         uuid.UUID
	eventType  string
	handlerKey string
	status     Status
	payload    []byte
	attempt    attempt
	eventDate  time.Time
}

// NewRecord creates new record that can be processed by inbox worker.
//
// Parameters:
//
//	id - is a unique id for inbox table. ID should be unique or storage
//			will ignore all duplicate ids.
//	eventType - is a topic with which event was published.
//	payload - the received body.
//	eventDate (optional) - when event was occurred.
func NewRecord(id uuid.UUID, eventType string, payload []byte, eventDate ...time.Time) (*Record, error) {
	if eventType == "" {
		return nil, fmt.Errorf("incorrect record event type provided")
	}

	date := time.Now().UTC()

	if len(eventDate) > 0 && !eventDate[0].IsZero() {
		date = eventDate[0]
	}

	return &Record{
		id:        id,
		eventType: eventType,
		payload:   payload,
		eventDate: date,
	}, nil
}

func newFullRecord(
	id uuid.UUID,
	status Status,
	eventType string,
	handlerKey string,
	payload []byte,
	currAttempt int,
	eventDate time.Time,
) *Record {
	return &Record{
		id:         id,
		status:     status,
		eventType:  eventType,
		handlerKey: handlerKey,
		payload:    payload,
		attempt: attempt{
			attempt: currAttempt,
		},
		eventDate: eventDate,
	}
}

// Done sets Done status to current Record. Status will be
// ignored on first save to the outbox table.
func (r *Record) Done() {
	r.status = Done
}

// Fail sets Failed status to current Record. Status will be
// ignored on first save to the outbox table.
func (r *Record) Fail(err error) {
	r.status = Failed

	r.attempt.message = err.Error()
	r.attempt.attempt++
}

func (r *Record) Dead() {
	r.status = Dead
}

// Null sets Null status to current Record.
func (r *Record) Null() {
	r.status = ""
}

func (r *Record) Attempt() int {
	return r.attempt.attempt
}

func (r *Record) CalcNewDeadline(dur time.Duration) {
	now := time.Now().UTC()
	now = now.Add(dur)

	r.attempt.nextAttempt = now
}

func (r *Record) withHandlerKey(key string) *Record {
	b := make([]byte, len(r.payload))
	copy(b, r.payload)

	return &Record{
		id:         r.id,
		eventType:  r.eventType,
		handlerKey: key,
		status:     r.status,
		payload:    b,
		eventDate:  r.eventDate,
	}
}
