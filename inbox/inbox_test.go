package inbox_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Melenium2/go-iobox/inbox"
)

func TestInbox_FailOrDead(t *testing.T) {
	svc := inbox.NewInbox(inbox.NewRegistry(), nil)

	t.Run("should fail next record", func(t *testing.T) {
		input := inbox.RecordWithAttempt(0, inbox.Progress)

		output := svc.FailOrDead(input, errors.New("err"))
		assert.Equal(t, 1, input.Attempt())
		assert.Equal(t, inbox.Failed, output.Status())
	})

	t.Run("should setup deadline for last attempt", func(t *testing.T) {
		input := inbox.RecordWithAttempt(3, inbox.Failed)

		output := svc.FailOrDead(input, errors.New("err"))
		assert.Equal(t, 4, output.Attempt())
		assert.Equal(t, inbox.Failed, output.Status())
		assert.Greater(t, output.Deadline(), time.Now().UTC())
	})

	t.Run("should mark record as 'dead'", func(t *testing.T) {
		input := inbox.RecordWithAttempt(4, inbox.Failed)

		output := svc.FailOrDead(input, errors.New("err"))
		assert.Equal(t, 5, output.Attempt())
		assert.Equal(t, inbox.Dead, output.Status())
	})
}
