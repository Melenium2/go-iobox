package outbox

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Melenium2/go-iobox/migration"
	"github.com/Melenium2/go-iobox/outbox/migrations"
)

// TODO объяснить что вынесли мигратор чтобы не нести
// зависимости коннекшена в сторадж и чтоб не блочить метрики.

type migrator struct {
	conn *sql.DB
}

func newMigrator(conn *sql.DB) *migrator {
	return &migrator{
		conn: conn,
	}
}

func (m *migrator) InitTable(ctx context.Context) error {
	migr := migration.New()

	if err := migr.SetupFS(ctx, m.conn, migrations.FS, "outbox_schema"); err != nil {
		return fmt.Errorf("failed to setup outbox migrations, %w", err)
	}

	err := migr.Up()
	if err == nil {
		return nil
	}

	_ = migr.Down()

	return fmt.Errorf("failed to run migrations, %w", err)
}
