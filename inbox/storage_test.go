package inbox_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq"

	"github.com/stretchr/testify/suite"

	"github.com/Melenium2/go-iobox/inbox"
)

type StorageSuite struct {
	suite.Suite

	db      *sql.DB
	storage *inbox.Storage
}

func (suite *StorageSuite) SetupSuite() {
	var (
		host     = "localhost"
		port     = "5437"
		user     = "postgres"
		pass     = "postgres"
		database = "outbox"
		address  = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, pass, database,
		)
	)

	db, err := sql.Open("postgres", address)
	suite.Require().NoError(err)

	err = db.Ping()
	suite.Require().NoError(err)

	suite.db = db
	suite.storage = inbox.NewStorage(db)

	err = suite.storage.InitInboxTable(context.Background())
	suite.Require().NoError(err)
}

func (suite *StorageSuite) SetupTest() {
}

func (suite *StorageSuite) TearDownTest() {
	truncateTable(suite.db)
}

func (suite *StorageSuite) TestFetch_Should_fetch_unprocessed_rows_and_set_new_status_to_rows() {
	initNotProcessedRows(suite.db)

	ctx := context.Background()

	expected := []*inbox.Record{
		inbox.Record1(),
		inbox.Record2(),
	}

	result, err := suite.storage.Fetch(ctx)
	suite.Require().NoError(err)
	suite.Assert().Equal(expected, result)
}

func (suite *StorageSuite) TestFetch_Should_no_fetch_rows_because_table_are_empty() {
	ctx := context.Background()

	truncateTable(suite.db)

	_, err := suite.storage.Fetch(ctx)
	suite.Require().ErrorIs(err, inbox.ErrNoRecords)
}

func (suite *StorageSuite) TestFetch_Should_no_fetch_rows_because_they_are_locked_by_another_operation() {
	initInProgressRows(suite.db)

	ctx := context.Background()

	_, err := suite.storage.Fetch(ctx)
	suite.Require().ErrorIs(err, inbox.ErrNoRecords)
}

func (suite *StorageSuite) TestFetch_Should_no_fetch_rows_if_all_rows_already_processed() {
	initDoneRows(suite.db)

	ctx := context.Background()

	_, err := suite.storage.Fetch(ctx)
	suite.Require().ErrorIs(err, inbox.ErrNoRecords)
}

func (suite *StorageSuite) TestUpdate_Should_update_provided_records_with_new_status() {
	initInProgressRows(suite.db)

	ctx := context.Background()

	record := inbox.Record1()
	record.Fail()

	err := suite.storage.Update(ctx, []*inbox.Record{record})
	suite.Require().NoError(err)

	{
		var (
			sqlStr   = "select status from __inbox_table where id = $1;"
			expected = "failed"
			dest     string
		)
		_ = suite.db.QueryRow(sqlStr, inbox.ID1()).Scan(&dest)
		suite.Assert().Equal(expected, dest)
	}
}

func (suite *StorageSuite) TestInsert_Should_insert_new_records_to_table() {
	initInProgressRows(suite.db)

	ctx := context.Background()

	newRecord, _ := inbox.NewRecord(inbox.ID3(), "3", []byte{})
	newRecord = newRecord.WithHandlerKey("3")

	err := suite.storage.Insert(ctx, newRecord)
	suite.Assert().NoError(err)

	{
		var (
			sqlStr         = "select status, event_type, handler_key, payload from __inbox_table where id = $1;"
			status         = sql.NullString{}
			eventType      = "3"
			handlerKey     = "3"
			payload        = []byte{}
			destStatus     sql.NullString
			destType       string
			destHandlerKey string
			destPayload    []byte
		)
		_ = suite.db.QueryRow(sqlStr, inbox.ID3()).Scan(&destStatus, &destType, &destHandlerKey, &destPayload)
		suite.Assert().Equal(status, destStatus)
		suite.Assert().Equal(eventType, destType)
		suite.Assert().Equal(payload, destPayload)
		suite.Assert().Equal(handlerKey, destHandlerKey)
	}
}

func (suite *StorageSuite) TestInsert_Should_not_insert_table_with_same_id_and_handler_key_already_existed_in_the_table() {
	initInProgressRows(suite.db)

	ctx := context.Background()

	newRecord, _ := inbox.NewRecord(inbox.ID1(), "1", []byte{})
	newRecord = newRecord.WithHandlerKey("1")

	err := suite.storage.Insert(ctx, newRecord)
	suite.Assert().NoError(err)

	{
		var (
			sqlStr    = "select count(id) from __inbox_table where id = $1;"
			count     = 1
			destCount int
		)
		_ = suite.db.QueryRow(sqlStr, inbox.ID1()).Scan(&destCount)
		suite.Assert().Equal(count, destCount)
	}
}

func TestStorageSuite(t *testing.T) {
	suite.Run(t, &StorageSuite{})
}

func truncateTable(db *sql.DB) {
	_, _ = db.Exec("delete from __inbox_table where id in ($1, $2, $3)", inbox.ID1(), inbox.ID2(), inbox.ID3())
}

func initNotProcessedRows(db *sql.DB) {
	_, _ = db.Exec(
		"insert into __inbox_table (id, event_type, handler_key, payload) values ($1, $2, $3, $4)",
		inbox.ID1(), "1", "1", "{}",
	)

	_, _ = db.Exec(
		"insert into __inbox_table (id, event_type, handler_key, payload) values ($1, $2, $3, $4)",
		inbox.ID2(), "1", "2", "{}",
	)

	_, _ = db.Exec(
		"insert into __inbox_table (id, status, event_type, handler_key, payload) values ($1, $2, $3, $4, $5)",
		inbox.ID3(), "done", "2", "1", "{}",
	)
}

func initInProgressRows(db *sql.DB) {
	_, _ = db.Exec(
		"insert into __inbox_table (id, status, event_type, handler_key, payload) values ($1, $2, $3, $4, $5)",
		inbox.ID1(), "progress", "1", "1", "{}",
	)

	_, _ = db.Exec(
		"insert into __inbox_table (id, status, event_type, handler_key, payload) values ($1, $2, $3, $4, $5)",
		inbox.ID2(), "progress", "1", "2", "{}",
	)
}

func initDoneRows(db *sql.DB) {
	_, _ = db.Exec(
		"insert into __inbox_table (id, status, event_type, handler_key, payload) values ($1, $2, $3, $4, $5)",
		inbox.ID1(), "done", "1", "1", "{}",
	)

	_, _ = db.Exec(
		"insert into __inbox_table (id, status, event_type, handler_key, payload) values ($1, $2, $3, $4, $5)",
		inbox.ID2(), "done", "1", "2", "{}",
	)
}
