package retention

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

var (
	DefaultEraseInterval   = 24 * time.Hour
	DefaultRetentionWindow = 60
)

type Logger interface {
	Print(...any)
	Printf(string, ...any)
}

type Config struct {
	EraseInterval   time.Duration
	RetentionWindow int
	Logger          Logger
}

func defaultConfig() Config {
	return Config{
		EraseInterval:   DefaultEraseInterval,
		RetentionWindow: DefaultRetentionWindow,
	}
}

type Policy struct {
	tableName string
	config    Config

	conn *sql.DB
}

func NewPolicy(conn *sql.DB, tableName string, config ...Config) *Policy {
	defaultConfig := defaultConfig()

	if len(config) > 0 && config[0].EraseInterval > 0 {
		defaultConfig.EraseInterval = config[0].EraseInterval
	}

	if len(config) > 0 && config[0].RetentionWindow > 0 {
		defaultConfig.RetentionWindow = config[0].RetentionWindow
	}

	return &Policy{
		conn:      conn,
		tableName: tableName,
		config:    defaultConfig,
	}
}

func (p *Policy) Start(ctx context.Context) {
	ticker := time.NewTicker(p.config.EraseInterval)

	for {
		select {
		case now := <-ticker.C:
			tailDate := now.AddDate(0, 0, p.config.RetentionWindow)

			deletedRows, err := p.erase(context.Background(), tailDate)
			if err != nil {
				// TODO: log error.
			}

			_ = deletedRows
		// TODO: log deleted rows.
		case <-ctx.Done():
			return
		}
	}
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
