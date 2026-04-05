package db

import (
	"fmt"
	"strings"

	"pornboss/internal/jav"
	"pornboss/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// Open initialises a GORM-backed SQLite database and applies schema migrations.
func Open(path string) (*gorm.DB, error) {
	driverName := registerSQLiteFunctions()
	db, err := gorm.Open(sqlite.New(sqlite.Config{DriverName: driverName, DSN: path}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Enable WAL mode for better concurrent read/write performance.
	if err := db.Exec("PRAGMA journal_mode=WAL;").Error; err != nil {
		return nil, fmt.Errorf("enable WAL: %w", err)
	}
	if err := db.Exec("PRAGMA foreign_keys=ON;").Error; err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	if err := db.AutoMigrate(
		&models.Directory{},
		&models.Video{},
		&models.Tag{},
		&models.VideoTag{},
		&models.Config{},
		&models.Jav{},
		&models.JavTag{},
		&models.JavIdol{},
		&models.JavTagMap{},
		&models.JavIdolMap{},
	); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}
	if err := db.Exec("UPDATE video SET play_count = 0 WHERE play_count IS NULL").Error; err != nil {
		return nil, fmt.Errorf("backfill play_count: %w", err)
	}
	if err := db.Exec("UPDATE jav SET provider = ? WHERE COALESCE(provider, 0) = 0", int(jav.ProviderJavBus)).Error; err != nil {
		return nil, fmt.Errorf("backfill jav provider: %w", err)
	}
	hasIsUser, err := sqliteColumnExists(db, "jav_tag", "is_user")
	if err != nil {
		return nil, fmt.Errorf("check jav_tag.is_user column: %w", err)
	}
	if hasIsUser {
		if err := db.Exec(
			"UPDATE jav_tag SET provider = CASE WHEN COALESCE(is_user, 0) = 1 THEN ? ELSE ? END WHERE COALESCE(provider, 0) = 0",
			int(jav.ProviderUser),
			int(jav.ProviderJavBus),
		).Error; err != nil {
			return nil, fmt.Errorf("backfill jav tag provider: %w", err)
		}
	} else {
		if err := db.Exec("UPDATE jav_tag SET provider = ? WHERE COALESCE(provider, 0) = 0", int(jav.ProviderJavBus)).Error; err != nil {
			return nil, fmt.Errorf("backfill jav tag provider: %w", err)
		}
	}
	if err := db.Exec("DROP INDEX IF EXISTS idx_jav_tag_name_source").Error; err != nil {
		return nil, fmt.Errorf("drop jav tag index: %w", err)
	}
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_jav_tag_name_source ON jav_tag(name, provider)").Error; err != nil {
		return nil, fmt.Errorf("create jav tag index: %w", err)
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_jav_tag_provider ON jav_tag(provider)").Error; err != nil {
		return nil, fmt.Errorf("create jav tag provider index: %w", err)
	}
	if hasIsUser {
		if err := db.Exec("ALTER TABLE jav_tag DROP COLUMN is_user").Error; err != nil && !ignorableSQLiteDropColumnErr(err) {
			return nil, fmt.Errorf("drop jav_tag.is_user: %w", err)
		}
	}
	if err := backfillJavIdolJapaneseNames(db); err != nil {
		return nil, fmt.Errorf("backfill jav idol japanese names: %w", err)
	}
	return db, nil
}

func backfillJavIdolJapaneseNames(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("nil db")
	}

	var idols []models.JavIdol
	if err := db.
		Where("(japanese_name IS NULL OR japanese_name = '') AND name IS NOT NULL AND name <> ''").
		Find(&idols).Error; err != nil {
		return err
	}

	ids := make([]int64, 0, len(idols))
	for _, idol := range idols {
		if looksLikeJapaneseName(idol.Name) {
			ids = append(ids, idol.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}

	return db.Model(&models.JavIdol{}).
		Where("id IN ?", ids).
		Where("japanese_name IS NULL OR japanese_name = ''").
		Update("japanese_name", gorm.Expr("name")).Error
}

func sqliteColumnExists(db *gorm.DB, table, column string) (bool, error) {
	rows, err := db.Raw(fmt.Sprintf("PRAGMA table_info(%s)", table)).Rows()
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid        int
			name       string
			colType    string
			notNull    int
			defaultV   any
			primaryKey int
		)
		if err := rows.Scan(&cid, &name, &colType, &notNull, &defaultV, &primaryKey); err != nil {
			return false, err
		}
		if strings.EqualFold(name, column) {
			return true, nil
		}
	}
	return false, rows.Err()
}

func ignorableSQLiteDropColumnErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "no such column") || strings.Contains(msg, "duplicate column name")
}

func looksLikeJapaneseName(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	for _, r := range value {
		switch {
		case r >= 0x3040 && r <= 0x30ff:
			return true
		case r >= 0x31f0 && r <= 0x31ff:
			return true
		case r >= 0x4e00 && r <= 0x9fff:
			return true
		case r >= 0xff66 && r <= 0xff9d:
			return true
		}
	}
	return false
}
