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

func nopCallback(err error) {}

type Config struct {
	EraseInterval   time.Duration
	RetentionWindow int
	ErrorCallback   func(err error)
}

func defaultConfig() Config {
	return Config{
		EraseInterval:   DefaultEraseInterval,
		RetentionWindow: DefaultRetentionWindow,
		ErrorCallback:   nopCallback,
	}
}

type Policy struct {
	tableName string
	config    Config

	conn *sql.DB
}

func NewPolicy(conn *sql.DB, tableName string, config ...Config) *Policy {
	cfg := defaultConfig()

	if len(config) > 0 && config[0].EraseInterval > 0 {
		cfg.EraseInterval = config[0].EraseInterval
	}

	if len(config) > 0 && config[0].RetentionWindow > 0 {
		cfg.RetentionWindow = config[0].RetentionWindow
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

func (p *Policy) Start(ctx context.Context) {
	ticker := time.NewTicker(p.config.EraseInterval)

	for {
		select {
		case now := <-ticker.C:
			tailDate := now.AddDate(0, 0, p.config.RetentionWindow)

			_, err := p.erase(context.Background(), tailDate)
			if err != nil {
				p.config.ErrorCallback(err)
			}
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
