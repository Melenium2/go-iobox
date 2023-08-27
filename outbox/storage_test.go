package outbox_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"

	"github.com/Melenium2/go-iobox/outbox"
)

type StorageSuite struct {
	suite.Suite

	db      *sql.DB
	storage *outbox.Storage
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
	suite.storage = outbox.NewStorage(db)

	err = suite.storage.InitOutboxTable(context.Background())
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

	expected := []*outbox.Record{
		outbox.Record1(),
		outbox.Record2(),
	}

	result, err := suite.storage.Fetch(ctx)
	suite.Require().NoError(err)
	suite.Assert().Equal(expected, result)
}

func (suite *StorageSuite) TestFetch_Should_no_fetch_rows_because_table_are_empty() {
	ctx := context.Background()

	truncateTable(suite.db)

	_, err := suite.storage.Fetch(ctx)
	suite.Require().ErrorIs(err, outbox.ErrNoRecrods)
}

func (suite *StorageSuite) TestFetch_Should_no_fetch_rows_because_they_are_locked_by_another_operation() {
	initInProgressRows(suite.db)

	ctx := context.Background()

	_, err := suite.storage.Fetch(ctx)
	suite.Require().ErrorIs(err, outbox.ErrNoRecrods)
}

func (suite *StorageSuite) TestFetch_Should_no_fetch_rows_if_all_rows_already_processed() {
	initDoneRows(suite.db)

	ctx := context.Background()

	_, err := suite.storage.Fetch(ctx)
	suite.Require().ErrorIs(err, outbox.ErrNoRecrods)
}

func (suite *StorageSuite) TestUpdate_Should_update_provided_records_with_new_status() {
	initInProgressRows(suite.db)

	ctx := context.Background()

	record := outbox.Record1()
	record.Fail()

	err := suite.storage.Update(ctx, []*outbox.Record{record})
	suite.Require().NoError(err)

	{
		var (
			sqlStr   = "select status from __outbox_table where id = $1;"
			expected = "failed"
			dest     string
		)
		_ = suite.db.QueryRow(sqlStr, outbox.ID1()).Scan(&dest)
		suite.Assert().Equal(expected, dest)
	}
}

func (suite *StorageSuite) TestUpdate_Should_set_null_status_to_record() {
	initInProgressRows(suite.db)

	ctx := context.Background()

	record1 := outbox.Record1()
	record1.Null()

	record2 := outbox.Record2()
	record2.Null()

	err := suite.storage.Update(ctx, []*outbox.Record{record1, record2})
	suite.Require().NoError(err)

	{
		var (
			sqlStr   = "select count(*) from __outbox_table where id in ($1, $2) and status is not null;"
			expected = 0
			dest     int
		)
		_ = suite.db.QueryRow(sqlStr, outbox.ID1(), outbox.ID2()).Scan(&dest)
		suite.Assert().Equal(expected, dest)
	}
}

func (suite *StorageSuite) TestInsert_Should_insert_new_records_to_table() {
	initInProgressRows(suite.db)

	ctx := context.Background()

	payload := outbox.PayloadMarshaler{Body: []byte("{}")}

	newRecord := outbox.NewRecord(outbox.ID3(), "topic2", &payload)

	err := suite.storage.Insert(ctx, suite.db, newRecord)
	suite.Assert().NoError(err)

	{
		var (
			sqlStr      = "select status, event_type, payload from __outbox_table where id = $1;"
			status      = sql.NullString{}
			eventType   = "topic2"
			payload     = []byte("{}")
			destStatus  sql.NullString
			destType    string
			destPayload []byte
		)
		_ = suite.db.QueryRow(sqlStr, outbox.ID3()).Scan(&destStatus, &destType, &destPayload)
		suite.Assert().Equal(status, destStatus)
		suite.Assert().Equal(eventType, destType)
		suite.Assert().Equal(payload, destPayload)
	}
}

func TestStorageSuite(t *testing.T) {
	suite.Run(t, &StorageSuite{})
}

func truncateTable(db *sql.DB) {
	_, _ = db.Exec("delete from __outbox_table where id in ($1, $2, $3)", outbox.ID1(), outbox.ID2(), outbox.ID3())
}

func initNotProcessedRows(db *sql.DB) {
	_, _ = db.Exec(
		"insert into __outbox_table (id, event_type, payload) values ($1, $2, $3)",
		outbox.ID1(), "topic1", "{}",
	)

	_, _ = db.Exec(
		"insert into __outbox_table (id, event_type, payload) values ($1, $2, $3)",
		outbox.ID2(), "topic1", "{}",
	)

	_, _ = db.Exec(
		"insert into __outbox_table (id, status, event_type, payload) values ($1, $2, $3, $4)",
		outbox.ID3(), "done", "topic1", "{}",
	)
}

func initInProgressRows(db *sql.DB) {
	_, _ = db.Exec(
		"insert into __outbox_table (id, status, event_type, payload) values ($1, $2, $3, $4)",
		outbox.ID1(), "progress", "topic1", "{}",
	)

	_, _ = db.Exec(
		"insert into __outbox_table (id, status, event_type, payload) values ($1, $2, $3, $4)",
		outbox.ID2(), "progress", "topic1", "{}",
	)
}

func initDoneRows(db *sql.DB) {
	_, _ = db.Exec(
		"insert into __outbox_table (id, status, event_type, payload) values ($1, $2, $3, $4)",
		outbox.ID1(), "done", "topic1", "{}",
	)

	_, _ = db.Exec(
		"insert into __outbox_table (id, status, event_type, payload) values ($1, $2, $3, $4)",
		outbox.ID2(), "done", "topic1", "{}",
	)
}
