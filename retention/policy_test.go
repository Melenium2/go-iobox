package retention

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const tableName = "__retention_table"

func insertWithDate(conn *sql.DB, date time.Time) error {
	sqlstr := "insert into " + tableName + " (created_at)  values ($1);"

	_, err := conn.Exec(sqlstr, date)

	return err
}

func createTable(conn *sql.DB) error {
	sqlstr := "create table if not exists " + tableName + " (" +
		"		created_at timestamp not null " +
		");"

	_, err := conn.Exec(sqlstr)

	return err
}

func dropTable(conn *sql.DB) error {
	sqlstr := "drop table if exists " + tableName + ";"

	_, err := conn.Exec(sqlstr)

	return err
}

func truncateTable(conn *sql.DB) error {
	sqlstr := "truncate table " + tableName + ";"

	_, err := conn.Exec(sqlstr)

	return err
}

type PolicySuite struct {
	suite.Suite

	db  *sql.DB
	svc *Policy
}

func TestPolicySuite(t *testing.T) {
	suite.Run(t, &PolicySuite{})
}

func (suite *PolicySuite) SetupSuite() {
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
	suite.svc = NewPolicy(db, tableName)

	err = createTable(suite.db)
	suite.Require().NoError(err)
}

func (suite *PolicySuite) TearDownTest() {
	err := truncateTable(suite.db)
	suite.Require().NoError(err)
}

func (suite *PolicySuite) TearDownSuite() {
	err := dropTable(suite.db)
	suite.Require().NoError(err)
}

func (suite *PolicySuite) TestErase_Should_remove_all_rows_where_created_at_less_then_specified_date() {
	ctx := context.Background()

	{
		var (
			date1 = time.Date(2007, 12, 2, 0, 0, 0, 0, time.UTC)
			date2 = time.Date(2007, 6, 1, 0, 0, 0, 0, time.UTC)
		)
		err := insertWithDate(suite.db, date1)
		suite.Require().NoError(err)
		err = insertWithDate(suite.db, date1)
		suite.Require().NoError(err)
		err = insertWithDate(suite.db, date2)
		suite.Require().NoError(err)
		err = insertWithDate(suite.db, date2)
		suite.Require().NoError(err)
		err = insertWithDate(suite.db, date2)
		suite.Require().NoError(err)
	}

	date := time.Date(2007, 6, 2, 0, 0, 0, 0, time.UTC)

	removed, err := suite.svc.erase(ctx, date)
	suite.Require().NoError(err)
	suite.Assert().Equal(int64(3), removed)
}

func (suite *PolicySuite) TestErase_Should_do_nothing_if_all_rows_are_newer_then_specified_date() {
	ctx := context.Background()

	{
		var (
			date1 = time.Date(2007, 12, 2, 0, 0, 0, 0, time.UTC)
			date2 = time.Date(2007, 6, 1, 0, 0, 0, 0, time.UTC)
		)
		err := insertWithDate(suite.db, date1)
		suite.Require().NoError(err)
		err = insertWithDate(suite.db, date1)
		suite.Require().NoError(err)
		err = insertWithDate(suite.db, date2)
		suite.Require().NoError(err)
		err = insertWithDate(suite.db, date2)
		suite.Require().NoError(err)
		err = insertWithDate(suite.db, date2)
		suite.Require().NoError(err)
	}

	date := time.Date(2007, 5, 2, 0, 0, 0, 0, time.UTC)

	removed, err := suite.svc.erase(ctx, date)
	suite.Require().NoError(err)
	suite.Assert().Equal(int64(0), removed)
}

func (suite *PolicySuite) TestErase_Should_do_nothing_if_no_rows_at_the_table() {
	ctx := context.Background()

	date := time.Date(2007, 5, 2, 0, 0, 0, 0, time.UTC)

	removed, err := suite.svc.erase(ctx, date)
	suite.Require().NoError(err)
	suite.Assert().Equal(int64(0), removed)
}

func TestTailDate(t *testing.T) {
	svc := &Policy{}

	t.Run("should calculate tail date after which we need to delete the rows", func(t *testing.T) {
		var (
			date       = time.Date(2006, 12, 1, 5, 5, 5, 0, time.UTC)
			windowDays = 60
		)

		expected := time.Date(2006, 10, 2, 5, 5, 5, 0, time.UTC)

		res := svc.tailDate(date, windowDays)
		assert.Equal(t, expected, res)
	})
}
