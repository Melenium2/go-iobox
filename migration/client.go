package migration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Client struct {
	once     sync.Once
	migrator *migrate.Migrate
}

func New() *Client {
	return &Client{}
}

func (c *Client) Setup(ctx context.Context, db *sql.DB, path, migrTable string) error {
	var setupErr error

	c.once.Do(func() {
		conn, err := db.Conn(ctx)
		if err != nil {
			setupErr = err

			return
		}

		cfg := &postgres.Config{
			MigrationsTable: migrTable,
		}

		post, err := postgres.WithConnection(ctx, conn, cfg)
		if err != nil {
			setupErr = err

			return
		}

		filePath := fmt.Sprintf("file://%s", path)

		migr, err := migrate.NewWithDatabaseInstance(filePath, "postgres", post)
		if err != nil {
			setupErr = err

			return
		}

		c.migrator = migr
	})

	return setupErr
}

func (c *Client) Up() error {
	if c.migrator == nil {
		return fmt.Errorf("migrate is not setup, use Setup() first")
	}

	err := c.migrator.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		return nil
	}

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Down() error {
	if c.migrator == nil {
		return fmt.Errorf("migrate is not setup, use Setup() first")
	}

	return c.migrator.Down()
}
