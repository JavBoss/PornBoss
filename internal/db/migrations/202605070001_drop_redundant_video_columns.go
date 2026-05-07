package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddNamedMigrationContext("202605070001_drop_redundant_video_columns.go", dropRedundantVideoColumns, irreversibleMigration)
}

func dropRedundantVideoColumns(ctx context.Context, tx *sql.Tx) error {
	if err := rebuildCanonicalTable(ctx, tx, videoContentTable); err != nil {
		return err
	}
	return execStatements(ctx, tx,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_video_fingerprint ON video(fingerprint)`,
	)
}

var videoContentTable = canonicalTable{
	name: "video",
	body: `(
		id integer PRIMARY KEY AUTOINCREMENT,
		size integer,
		fingerprint text,
		duration_sec integer,
		play_count integer NOT NULL DEFAULT 0,
		created_at datetime,
		updated_at datetime
	)`,
	columns: columns(
		"id", "integer",
		"size", "integer",
		"fingerprint", "text",
		"duration_sec", "integer",
		"play_count", "integer NOT NULL DEFAULT 0",
		"created_at", "datetime",
		"updated_at", "datetime",
	),
}
