package migrations

import (
	"context"
	"database/sql"
	"strings"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddNamedMigrationContext("202605060008_migrate_to_canonical_schema.go", migrateToCanonicalSchema, irreversibleMigration)
}

func migrateToCanonicalSchema(ctx context.Context, tx *sql.Tx) error {
	if err := ensureCurrentTables(ctx, tx); err != nil {
		return err
	}
	if err := ensureCurrentColumns(ctx, tx); err != nil {
		return err
	}
	if err := backfillLegacyData(ctx, tx); err != nil {
		return err
	}
	if err := rebuildCanonicalTables(ctx, tx); err != nil {
		return err
	}
	return createCurrentIndexes(ctx, tx)
}

func ensureCurrentTables(ctx context.Context, tx *sql.Tx) error {
	for _, table := range currentTables {
		if err := execDB(ctx, tx, table.createSQL(table.name, true)); err != nil {
			return err
		}
	}
	return nil
}

func ensureCurrentColumns(ctx context.Context, tx *sql.Tx) error {
	for _, table := range currentTables {
		for _, col := range table.columns {
			if err := addColumnIfMissing(ctx, tx, table.name, col.name, col.decl); err != nil {
				return err
			}
		}
	}
	return nil
}

func backfillLegacyData(ctx context.Context, tx *sql.Tx) error {
	if err := execStatements(ctx, tx,
		`UPDATE video SET play_count = 0 WHERE play_count IS NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_video_location_directory_path ON video_location(directory_id, relative_path)`,
	); err != nil {
		return err
	}
	if err := backfillVideoLocations(ctx, tx); err != nil {
		return err
	}
	if err := backfillJavProviders(ctx, tx); err != nil {
		return err
	}
	return backfillJavIdols(ctx, tx)
}

func backfillVideoLocations(ctx context.Context, tx *sql.Tx) error {
	hasLocations, err := rowsExist(ctx, tx, `SELECT 1 FROM video_location LIMIT 1`)
	if err != nil {
		return err
	}
	if !hasLocations {
		if err := execStatements(ctx, tx,
			`INSERT OR IGNORE INTO video_location (
				video_id,
				directory_id,
				relative_path,
				filename,
				modified_at,
				jav_id,
				is_delete,
				created_at,
				updated_at
			)
			SELECT
				id,
				directory_id,
				path,
				filename,
				modified_at,
				jav_id,
				COALESCE(hidden, 0),
				created_at,
				updated_at
			FROM video
			WHERE directory_id > 0 AND COALESCE(path, '') <> ''`,
			`UPDATE video_location
			 SET jav_id = (
				SELECT video.jav_id
				FROM video
				WHERE video.id = video_location.video_id
			 )
			 WHERE jav_id IS NULL
			   AND EXISTS (
				SELECT 1
				FROM video
				WHERE video.id = video_location.video_id
				  AND video.jav_id IS NOT NULL
			   )`,
		); err != nil {
			return err
		}
	}
	return backfillVideoLocationFilenames(ctx, tx)
}

func backfillVideoLocationFilenames(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, relative_path
		FROM video_location
		WHERE COALESCE(filename, '') = '' AND COALESCE(relative_path, '') <> ''
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type location struct {
		id           int64
		relativePath string
	}
	var locations []location
	for rows.Next() {
		var loc location
		if err := rows.Scan(&loc.id, &loc.relativePath); err != nil {
			return err
		}
		locations = append(locations, loc)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, loc := range locations {
		filename := baseNameFromSlashPath(loc.relativePath)
		if filename == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `UPDATE video_location SET filename = ? WHERE id = ?`, filename, loc.id); err != nil {
			return err
		}
	}
	return nil
}

func backfillJavProviders(ctx context.Context, tx *sql.Tx) error {
	hasIsUser, err := columnExists(ctx, tx, "jav_tag", "is_user")
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jav SET provider = ? WHERE COALESCE(provider, 0) = 0`, providerJavBus); err != nil {
		return err
	}
	if hasIsUser {
		_, err = tx.ExecContext(ctx,
			`UPDATE jav_tag
			 SET provider = CASE WHEN COALESCE(is_user, 0) = 1 THEN ? ELSE ? END
			 WHERE COALESCE(provider, 0) = 0`,
			providerUser,
			providerJavBus,
		)
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE jav_tag SET provider = ? WHERE COALESCE(provider, 0) = 0`, providerJavBus)
	return err
}

func backfillJavIdols(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, `
		UPDATE jav_idol
		SET is_english = 1
		WHERE COALESCE(is_english, 0) = 0
		  AND id IN (
			SELECT jim.jav_idol_id
			FROM jav_idol_map jim
			JOIN jav j ON j.id = jim.jav_id
			WHERE j.provider = ?
		  )
	`, providerJavDatabase); err != nil {
		return err
	}
	return backfillJavIdolJapaneseNames(ctx, tx)
}

func backfillJavIdolJapaneseNames(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, name
		FROM jav_idol
		WHERE (japanese_name IS NULL OR japanese_name = '') AND name IS NOT NULL AND name <> ''
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type idol struct {
		id   int64
		name string
	}
	var idols []idol
	for rows.Next() {
		var item idol
		if err := rows.Scan(&item.id, &item.name); err != nil {
			return err
		}
		idols = append(idols, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, item := range idols {
		if !looksLikeJapaneseName(item.name) {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE jav_idol
			SET japanese_name = name
			WHERE id = ? AND (japanese_name IS NULL OR japanese_name = '')
		`, item.id); err != nil {
			return err
		}
	}
	return nil
}

func rebuildCanonicalTables(ctx context.Context, tx *sql.Tx) error {
	for _, table := range currentTables {
		if err := rebuildCanonicalTable(ctx, tx, table); err != nil {
			return err
		}
	}
	return nil
}

func rebuildCanonicalTable(ctx context.Context, tx *sql.Tx, table canonicalTable) error {
	newName := "__new_" + table.name
	if err := execStatements(ctx, tx,
		`DROP TABLE IF EXISTS `+quoteIdentifier(newName),
		table.createSQL(newName, false),
	); err != nil {
		return err
	}
	columns := table.quotedColumns()
	if err := execDB(ctx, tx,
		`INSERT INTO `+quoteIdentifier(newName)+` (`+columns+`)
		 SELECT `+columns+` FROM `+quoteIdentifier(table.name),
	); err != nil {
		return err
	}
	return execStatements(ctx, tx,
		`DROP TABLE `+quoteIdentifier(table.name),
		`ALTER TABLE `+quoteIdentifier(newName)+` RENAME TO `+quoteIdentifier(table.name),
	)
}

func createCurrentIndexes(ctx context.Context, tx *sql.Tx) error {
	return execStatements(ctx, tx,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_directory_path ON directory(path)`,
		`CREATE INDEX IF NOT EXISTS idx_directory_missing ON directory(missing)`,
		`CREATE INDEX IF NOT EXISTS idx_directory_is_delete ON directory(is_delete)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_jav_code ON jav(code)`,
		`CREATE INDEX IF NOT EXISTS idx_video_directory_id ON video(directory_id)`,
		`CREATE INDEX IF NOT EXISTS idx_video_path ON video(path)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_video_fingerprint ON video(fingerprint)`,
		`DROP INDEX IF EXISTS idx_video_jav_id_visible`,
		`CREATE INDEX IF NOT EXISTS idx_video_jav_id ON video(jav_id)`,
		`CREATE INDEX IF NOT EXISTS idx_video_hidden ON video(hidden)`,
		`CREATE INDEX IF NOT EXISTS idx_video_location_video_id ON video_location(video_id)`,
		`CREATE INDEX IF NOT EXISTS idx_video_location_directory_id ON video_location(directory_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_video_location_directory_path ON video_location(directory_id, relative_path)`,
		`CREATE INDEX IF NOT EXISTS idx_video_location_filename ON video_location(filename)`,
		`CREATE INDEX IF NOT EXISTS idx_video_location_jav_id ON video_location(jav_id)`,
		`CREATE INDEX IF NOT EXISTS idx_video_location_is_delete ON video_location(is_delete)`,
		`CREATE INDEX IF NOT EXISTS idx_video_location_jav_id_is_delete ON video_location(jav_id, is_delete)`,
		`CREATE INDEX IF NOT EXISTS idx_video_location_video_id_jav_id ON video_location(video_id, jav_id)`,
		`CREATE INDEX IF NOT EXISTS idx_video_location_visible_path ON video_location(jav_id, is_delete, relative_path)`,
		`CREATE INDEX IF NOT EXISTS idx_video_location_visible_filename ON video_location(jav_id, is_delete, filename COLLATE NOCASE)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_tag_name ON tag(name)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_jav_tag_name_source ON jav_tag(name, provider)`,
		`CREATE INDEX IF NOT EXISTS idx_jav_tag_provider ON jav_tag(provider)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_jav_idol_name_language ON jav_idol(name, is_english)`,
		`CREATE INDEX IF NOT EXISTS idx_jav_idol_map_jav_idol_id_jav_id ON jav_idol_map(jav_idol_id, jav_id)`,
	)
}

type canonicalTable struct {
	name    string
	body    string
	columns []canonicalColumn
}

type canonicalColumn struct {
	name string
	decl string
}

func (t canonicalTable) createSQL(name string, ifNotExists bool) string {
	optional := ""
	if ifNotExists {
		optional = " IF NOT EXISTS"
	}
	return "CREATE TABLE" + optional + " " + quoteIdentifier(name) + " " + t.body
}

func (t canonicalTable) quotedColumns() string {
	columns := make([]string, 0, len(t.columns))
	for _, col := range t.columns {
		columns = append(columns, quoteIdentifier(col.name))
	}
	return strings.Join(columns, ", ")
}

var currentTables = []canonicalTable{
	{
		name: "directory",
		body: `(
			id integer PRIMARY KEY AUTOINCREMENT,
			path text,
			missing numeric,
			is_delete numeric,
			created_at datetime,
			updated_at datetime
		)`,
		columns: columns(
			"id", "integer",
			"path", "text",
			"missing", "numeric",
			"is_delete", "numeric",
			"created_at", "datetime",
			"updated_at", "datetime",
		),
	},
	{
		name: "jav",
		body: `(
			id integer PRIMARY KEY AUTOINCREMENT,
			code text,
			title text,
			release_unix integer,
			duration_min integer,
			provider integer NOT NULL DEFAULT 0,
			fetched_at datetime,
			created_at datetime,
			updated_at datetime
		)`,
		columns: columns(
			"id", "integer",
			"code", "text",
			"title", "text",
			"release_unix", "integer",
			"duration_min", "integer",
			"provider", "integer NOT NULL DEFAULT 0",
			"fetched_at", "datetime",
			"created_at", "datetime",
			"updated_at", "datetime",
		),
	},
	{
		name: "video",
		body: `(
			id integer PRIMARY KEY AUTOINCREMENT,
			directory_id integer NOT NULL,
			path text,
			filename text,
			size integer,
			modified_at datetime,
			fingerprint text,
			duration_sec integer,
			play_count integer NOT NULL DEFAULT 0,
			created_at datetime,
			updated_at datetime,
			jav_id integer,
			hidden numeric,
			CONSTRAINT fk_video_directory FOREIGN KEY (directory_id) REFERENCES directory(id) ON UPDATE CASCADE ON DELETE RESTRICT,
			CONSTRAINT fk_video_jav FOREIGN KEY (jav_id) REFERENCES jav(id) ON UPDATE CASCADE ON DELETE SET NULL
		)`,
		columns: columns(
			"id", "integer",
			"directory_id", "integer NOT NULL DEFAULT 0",
			"path", "text",
			"filename", "text",
			"size", "integer",
			"modified_at", "datetime",
			"fingerprint", "text",
			"duration_sec", "integer",
			"play_count", "integer NOT NULL DEFAULT 0",
			"created_at", "datetime",
			"updated_at", "datetime",
			"jav_id", "integer",
			"hidden", "numeric",
		),
	},
	{
		name: "video_location",
		body: `(
			id integer PRIMARY KEY AUTOINCREMENT,
			video_id integer NOT NULL,
			directory_id integer NOT NULL,
			relative_path text NOT NULL,
			filename text,
			modified_at datetime,
			jav_id integer,
			is_delete numeric,
			created_at datetime,
			updated_at datetime,
			CONSTRAINT fk_video_location_video FOREIGN KEY (video_id) REFERENCES video(id) ON UPDATE CASCADE ON DELETE CASCADE,
			CONSTRAINT fk_video_location_directory FOREIGN KEY (directory_id) REFERENCES directory(id) ON UPDATE CASCADE ON DELETE RESTRICT
		)`,
		columns: columns(
			"id", "integer",
			"video_id", "integer NOT NULL DEFAULT 0",
			"directory_id", "integer NOT NULL DEFAULT 0",
			"relative_path", "text NOT NULL DEFAULT ''",
			"filename", "text",
			"modified_at", "datetime",
			"jav_id", "integer",
			"is_delete", "numeric",
			"created_at", "datetime",
			"updated_at", "datetime",
		),
	},
	{
		name: "tag",
		body: `(
			id integer PRIMARY KEY AUTOINCREMENT,
			name text,
			created_at datetime,
			updated_at datetime
		)`,
		columns: columns(
			"id", "integer",
			"name", "text",
			"created_at", "datetime",
			"updated_at", "datetime",
		),
	},
	{
		name: "video_tag",
		body: `(
			video_id integer,
			tag_id integer,
			created_at datetime,
			PRIMARY KEY (video_id, tag_id),
			CONSTRAINT fk_video_tag_video FOREIGN KEY (video_id) REFERENCES video(id) ON UPDATE CASCADE ON DELETE CASCADE,
			CONSTRAINT fk_video_tag_tag FOREIGN KEY (tag_id) REFERENCES tag(id) ON UPDATE CASCADE ON DELETE CASCADE
		)`,
		columns: columns(
			"video_id", "integer",
			"tag_id", "integer",
			"created_at", "datetime",
		),
	},
	{
		name: "config",
		body: `(
			key text PRIMARY KEY,
			value text,
			created_at datetime,
			updated_at datetime
		)`,
		columns: columns(
			"key", "text",
			"value", "text",
			"created_at", "datetime",
			"updated_at", "datetime",
		),
	},
	{
		name: "jav_tag",
		body: `(
			id integer PRIMARY KEY AUTOINCREMENT,
			name text,
			provider integer NOT NULL DEFAULT 0,
			created_at datetime,
			updated_at datetime
		)`,
		columns: columns(
			"id", "integer",
			"name", "text",
			"provider", "integer NOT NULL DEFAULT 0",
			"created_at", "datetime",
			"updated_at", "datetime",
		),
	},
	{
		name: "jav_idol",
		body: `(
			id integer PRIMARY KEY AUTOINCREMENT,
			name text,
			is_english numeric NOT NULL DEFAULT 0,
			roman_name text,
			japanese_name text,
			chinese_name text,
			height_cm integer,
			birth_date datetime,
			bust integer,
			waist integer,
			hips integer,
			cup integer,
			created_at datetime,
			updated_at datetime
		)`,
		columns: columns(
			"id", "integer",
			"name", "text",
			"is_english", "numeric NOT NULL DEFAULT 0",
			"roman_name", "text",
			"japanese_name", "text",
			"chinese_name", "text",
			"height_cm", "integer",
			"birth_date", "datetime",
			"bust", "integer",
			"waist", "integer",
			"hips", "integer",
			"cup", "integer",
			"created_at", "datetime",
			"updated_at", "datetime",
		),
	},
	{
		name: "jav_tag_map",
		body: `(
			jav_id integer,
			jav_tag_id integer,
			created_at datetime,
			PRIMARY KEY (jav_id, jav_tag_id),
			CONSTRAINT fk_jav_tag_map_jav FOREIGN KEY (jav_id) REFERENCES jav(id) ON UPDATE CASCADE ON DELETE CASCADE,
			CONSTRAINT fk_jav_tag_map_jav_tag FOREIGN KEY (jav_tag_id) REFERENCES jav_tag(id) ON UPDATE CASCADE ON DELETE CASCADE
		)`,
		columns: columns(
			"jav_id", "integer",
			"jav_tag_id", "integer",
			"created_at", "datetime",
		),
	},
	{
		name: "jav_idol_map",
		body: `(
			jav_id integer,
			jav_idol_id integer,
			created_at datetime,
			PRIMARY KEY (jav_id, jav_idol_id),
			CONSTRAINT fk_jav_idol_map_jav FOREIGN KEY (jav_id) REFERENCES jav(id) ON UPDATE CASCADE ON DELETE CASCADE,
			CONSTRAINT fk_jav_idol_map_jav_idol FOREIGN KEY (jav_idol_id) REFERENCES jav_idol(id) ON UPDATE CASCADE ON DELETE CASCADE
		)`,
		columns: columns(
			"jav_id", "integer",
			"jav_idol_id", "integer",
			"created_at", "datetime",
		),
	},
}

func columns(values ...string) []canonicalColumn {
	cols := make([]canonicalColumn, 0, len(values)/2)
	for i := 0; i+1 < len(values); i += 2 {
		cols = append(cols, canonicalColumn{name: values[i], decl: values[i+1]})
	}
	return cols
}
