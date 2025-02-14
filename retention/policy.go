package retention

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

var (
	// DefaultEraseInterval is default interval for next iteration
	// of erasing old data.
	DefaultEraseInterval = 24 * time.Hour
	// DefaultRetentionWindow is the interval of days that will
	// not be deleted. All data older than current number of days
	// will be deleted.
	DefaultRetentionWindow = 60
)

func nopCallback(err error) {}

type Config struct {
	// Interval for the next iteratin of erasing old data.
	//
	// Optional. By default: DefaultEraseInterval.
	EraseInterval time.Duration
	// The data older then current number of days will be deleted
	// at the next iteration.
	//
	// Optional. Default: DefaultRetentionWindow.
	RetentionWindowDays int
	// Callback to handle an error if one occurs while erasing data.
	//
	// Optional.
	ErrorCallback func(err error)
}

func defaultConfig() Config {
	return Config{
		EraseInterval:       DefaultEraseInterval,
		RetentionWindowDays: DefaultRetentionWindow,
		ErrorCallback:       nopCallback,
	}
}

// Policy is a structure that deletes data older than specified interval.
// This is useful if we do not need to keep all old data all the time.
type Policy struct {
	tableName string
	config    Config

	conn *sql.DB
}

// NewPolicy creates new Policy to erasing old data at the specified table.
//
// Arguments.
//
//	conn - connection to the database we need to interact with.
//	tableName - the name of the table in which we need to save the data window only.
//	config - optional configuration of retention policy.
func NewPolicy(conn *sql.DB, tableName string, config ...Config) *Policy {
	cfg := defaultConfig()

	if len(config) > 0 && config[0].EraseInterval > 0 {
		cfg.EraseInterval = config[0].EraseInterval
	}

	if len(config) > 0 && config[0].RetentionWindowDays > 0 {
		cfg.RetentionWindowDays = config[0].RetentionWindowDays
	}

	if len(config) > 0 && config[0].ErrorCallback != nil {
		cfg.ErrorCallback = config[0].ErrorCallback
	}

	return &Policy{
		conn:      conn,
		tableName: tableName,
		config:    cfg,
	}
}

// Start retention process. The function is blocking the main loop.
// Close the context to stop erasing process.
func (p *Policy) Start(ctx context.Context) {
	ticker := time.NewTicker(p.config.EraseInterval)

	for {
		select {
		case now := <-ticker.C:
			tailDate := p.tailDate(now, p.config.RetentionWindowDays)

			if err := p.iteration(tailDate); err != nil {
				p.config.ErrorCallback(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (p *Policy) tailDate(now time.Time, window int) time.Time {
	return now.AddDate(0, 0, -window)
}

func (p *Policy) iteration(tailDate time.Time) error {
	ctx := context.Background()

	_, err := p.erase(ctx, tailDate)

	return err
}

func (p *Policy) erase(ctx context.Context, tailDate time.Time) (int64, error) {
	sqlstr := "delete from " + p.tableName + " where created_at::date < $1::date;"

	result, err := p.conn.ExecContext(ctx, sqlstr, tailDate)
	if err != nil {
		return 0, fmt.Errorf("erase sql not executed, %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affected, nil
}
