package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddNamedMigrationContext("202605070002_add_jav_title_en.go", addJavTitleEn, irreversibleMigration)
}

func addJavTitleEn(ctx context.Context, tx *sql.Tx) error {
	if err := addColumnIfMissing(ctx, tx, "jav", "title_en", "text"); err != nil {
		return err
	}
	if err := execDB(ctx, tx,
		`UPDATE jav
		 SET title_en = title,
		     title = ''
		 WHERE provider = ?
		   AND COALESCE(title_en, '') = ''
		   AND COALESCE(title, '') <> ''`,
		providerJavDatabase,
	); err != nil {
		return err
	}
	return rebuildJavTableWithTitleEn(ctx, tx)
}

func rebuildJavTableWithTitleEn(ctx context.Context, tx *sql.Tx) error {
	const columns = `"id", "code", "title", "title_en", "release_unix", "duration_min", "provider", "fetched_at", "created_at", "updated_at"`
	if err := execStatements(ctx, tx,
		`DROP TABLE IF EXISTS "__new_jav"`,
		`CREATE TABLE "__new_jav" (
			id integer PRIMARY KEY AUTOINCREMENT,
			code text,
			title text,
			title_en text,
			release_unix integer,
			duration_min integer,
			provider integer NOT NULL DEFAULT 0,
			fetched_at datetime,
			created_at datetime,
			updated_at datetime
		)`,
		`INSERT INTO "__new_jav" (`+columns+`)
		 SELECT `+columns+` FROM "jav"`,
		`DROP TABLE "jav"`,
		`ALTER TABLE "__new_jav" RENAME TO "jav"`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_jav_code ON jav(code)`,
	); err != nil {
		return err
	}
	return nil
}
