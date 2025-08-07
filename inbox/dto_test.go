package inbox_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Melenium2/go-iobox/inbox"
)

func TestMakeRecords_Should_sorts_dtos_by_created_at_in_asc_order(t *testing.T) {
	dtos := []*inbox.DTORecord{
		{
			ID:         inbox.ID2().String(),
			Status:     "progress",
			EventType:  "1",
			HandlerKey: "2",
			Payload:    []byte("{}"),
			CreatedAt:  time.Date(2024, 6, 5, 17, 55, 1, 0, time.UTC),
		},
		{
			ID:         inbox.ID1().String(),
			Status:     "progress",
			EventType:  "1",
			HandlerKey: "1",
			Payload:    []byte("{}"),
			CreatedAt:  time.Date(2024, 6, 5, 17, 55, 2, 0, time.UTC),
		},
		{
			ID:         inbox.ID3().String(),
			Status:     "",
			EventType:  "2",
			HandlerKey: "1",
			Payload:    []byte("{}"),
			CreatedAt:  time.Date(2024, 6, 5, 17, 55, 15, 0, time.UTC),
		},
	}

	expected := []*inbox.Record{
		inbox.Record2(),
		inbox.Record1(),
		inbox.Record3(),
	}

	res, err := inbox.MakeRecords(dtos)
	require.NoError(t, err)
	assert.Equal(t, expected, res)
}
