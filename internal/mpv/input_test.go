package mpv

import (
	"context"
	"strings"
	"testing"

	"pornboss/internal/common"
	dbpkg "pornboss/internal/db"
)

func TestBuildConfigContentIncludesRequiredDefaults(t *testing.T) {
	prevDB := common.DB
	common.DB = nil
	defer func() {
		common.DB = prevDB
	}()

	content, err := buildConfigContent()
	if err != nil {
		t.Fatalf("buildConfigContent returned error: %v", err)
	}

	if !strings.Contains(content, "keep-open=yes\n") {
		t.Fatalf("expected keep-open=yes in mpv config, got %q", content)
	}
	if !strings.Contains(content, "ontop=yes\n") {
		t.Fatalf("expected ontop=yes in mpv config, got %q", content)
	}
}

func TestBuildConfigContentRespectsConfiguredOntop(t *testing.T) {
	openConfigTestDB(t)
	if err := dbpkg.UpsertConfig(context.Background(), map[string]string{
		playerOntopConfigKey: "false",
	}); err != nil {
		t.Fatalf("upsert config: %v", err)
	}

	content, err := buildConfigContent()
	if err != nil {
		t.Fatalf("buildConfigContent returned error: %v", err)
	}

	if !strings.Contains(content, "ontop=no\n") {
		t.Fatalf("expected ontop=no in mpv config, got %q", content)
	}
}

func openConfigTestDB(t *testing.T) {
	t.Helper()

	prevDB := common.DB
	db, err := dbpkg.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	common.DB = db
	t.Cleanup(func() {
		common.DB = prevDB
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

}
