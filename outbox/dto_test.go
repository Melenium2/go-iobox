package outbox_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Melenium2/go-iobox/outbox"
)

func TestMakeRecords_Should_sorts_dtos_by_created_at_in_asc_order(t *testing.T) {
	dtos := []*outbox.DTORecord{
		{
			ID:        outbox.ID2().String(),
			Status:    "progress",
			EventType: "topic1",
			Payload:   []byte("{}"),
			CreatedAt: time.Date(2000, 1, 1, 1, 15, 0, 0, time.UTC),
		},
		{
			ID:        outbox.ID1().String(),
			Status:    "progress",
			EventType: "topic1",
			Payload:   []byte("{}"),
			CreatedAt: time.Date(2000, 1, 1, 1, 13, 0, 0, time.UTC),
		},
		{
			ID:        outbox.ID3().String(),
			Status:    "done",
			EventType: "topic1",
			Payload:   []byte("{}"),
			CreatedAt: time.Date(2000, 1, 1, 1, 17, 0, 0, time.UTC),
		},
	}

	expected := []*outbox.Record{
		outbox.Record1(),
		outbox.Record2(),
		outbox.Record3(),
	}

	res, err := outbox.MakeRecrods(dtos)
	require.NoError(t, err)
	assert.Equal(t, expected, res)
}
