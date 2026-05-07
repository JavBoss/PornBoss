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
	if err := rebuildVideoContentTable(ctx, tx); err != nil {
		return err
	}
	return execStatements(ctx, tx,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_video_fingerprint ON video(fingerprint)`,
	)
}

func rebuildVideoContentTable(ctx context.Context, tx *sql.Tx) error {
	const columns = `"id", "size", "fingerprint", "duration_sec", "play_count", "created_at", "updated_at"`
	return execStatements(ctx, tx,
		`DROP TABLE IF EXISTS "__new_video"`,
		`CREATE TABLE "__new_video" (
			id integer PRIMARY KEY AUTOINCREMENT,
			size integer,
			fingerprint text,
			duration_sec integer,
			play_count integer NOT NULL DEFAULT 0,
			created_at datetime,
			updated_at datetime
		)`,
		`INSERT INTO "__new_video" (`+columns+`)
		 SELECT `+columns+` FROM "video"`,
		`DROP TABLE "video"`,
		`ALTER TABLE "__new_video" RENAME TO "video"`,
	)
}
