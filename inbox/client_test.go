package inbox_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Melenium2/go-iobox/inbox"
	"github.com/Melenium2/go-iobox/inbox/mocks"
)

func TestClient_WriteInbox_Should_insert_record_with_single_event_type_and_handler_key(t *testing.T) {
	ctx := context.Background()

	fakeConn := &mocks.SQLConn{}
	fakeConn.
		On("ExecContext", append([]any{ctx, mock.IsType("sql string")}, inbox.Record1Values()...)...).
		Return(nil, nil)

	fakeHandler := mocks.NewHandler(t)
	fakeHandler.On("Key").Return("1")

	handlerMap := map[string][]inbox.Handler{
		"1": {fakeHandler},
	}

	client := inbox.NewClient(inbox.NewStorage(fakeConn), handlerMap)

	err := client.WriteInbox(ctx, inbox.Record1())
	assert.NoError(t, err)
}
