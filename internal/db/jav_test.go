package db

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"pornboss/internal/common"
	"pornboss/internal/models"

	"gorm.io/gorm"
)

func TestListJavIdolsOnlyIncludesIdolsWithVisibleSoloWorks(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	now := time.Unix(1710000000, 0).UTC()

	dir := models.Directory{Path: "/tmp/media"}
	if err := db.Create(&dir).Error; err != nil {
		t.Fatalf("create directory: %v", err)
	}

	soloIdol := models.JavIdol{Name: "Solo Idol"}
	groupOnlyIdol := models.JavIdol{Name: "Group Only Idol"}
	if err := db.Create(&soloIdol).Error; err != nil {
		t.Fatalf("create solo idol: %v", err)
	}
	if err := db.Create(&groupOnlyIdol).Error; err != nil {
		t.Fatalf("create group idol: %v", err)
	}

	soloJav := models.Jav{Code: "AAA-001", Title: "Solo Work", Provider: 1, FetchedAt: now}
	groupJav := models.Jav{Code: "BBB-001", Title: "Group Work", Provider: 1, FetchedAt: now}
	hiddenSoloJav := models.Jav{Code: "CCC-001", Title: "Hidden Solo Work", Provider: 1, FetchedAt: now}
	if err := db.Create(&soloJav).Error; err != nil {
		t.Fatalf("create solo jav: %v", err)
	}
	if err := db.Create(&groupJav).Error; err != nil {
		t.Fatalf("create group jav: %v", err)
	}
	if err := db.Create(&hiddenSoloJav).Error; err != nil {
		t.Fatalf("create hidden solo jav: %v", err)
	}

	maps := []models.JavIdolMap{
		{JavID: soloJav.ID, JavIdolID: soloIdol.ID},
		{JavID: groupJav.ID, JavIdolID: soloIdol.ID},
		{JavID: groupJav.ID, JavIdolID: groupOnlyIdol.ID},
		{JavID: hiddenSoloJav.ID, JavIdolID: groupOnlyIdol.ID},
	}
	if err := db.Create(&maps).Error; err != nil {
		t.Fatalf("create idol maps: %v", err)
	}

	videos := []models.Video{
		{
			DirectoryID: dir.ID,
			Path:        "solo.mp4",
			Filename:    "solo.mp4",
			Fingerprint: "fp-solo",
			JavID:       int64Ptr(soloJav.ID),
			ModifiedAt:  now,
		},
		{
			DirectoryID: dir.ID,
			Path:        "group.mp4",
			Filename:    "group.mp4",
			Fingerprint: "fp-group",
			JavID:       int64Ptr(groupJav.ID),
			ModifiedAt:  now,
		},
		{
			DirectoryID: dir.ID,
			Path:        "hidden.mp4",
			Filename:    "hidden.mp4",
			Fingerprint: "fp-hidden",
			JavID:       int64Ptr(hiddenSoloJav.ID),
			ModifiedAt:  now,
			Hidden:      true,
		},
	}
	if err := db.Create(&videos).Error; err != nil {
		t.Fatalf("create videos: %v", err)
	}

	items, total, err := ListJavIdols(ctx, "", "", 20, 0)
	if err != nil {
		t.Fatalf("ListJavIdols: %v", err)
	}

	if total != 1 {
		t.Fatalf("unexpected total: got %d want 1", total)
	}
	if len(items) != 1 {
		t.Fatalf("unexpected item count: got %d want 1", len(items))
	}
	if items[0].ID != soloIdol.ID {
		t.Fatalf("unexpected idol id: got %d want %d", items[0].ID, soloIdol.ID)
	}
	if items[0].WorkCount != 2 {
		t.Fatalf("unexpected work count: got %d want 2", items[0].WorkCount)
	}
	if items[0].SampleCode != soloJav.Code {
		t.Fatalf("unexpected sample code: got %q want %q", items[0].SampleCode, soloJav.Code)
	}
}

func TestGetJavIdolSummaryReturnsSampleCodeAndWorkCount(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	now := time.Unix(1710000000, 0).UTC()

	dir := models.Directory{Path: "/tmp/media"}
	if err := db.Create(&dir).Error; err != nil {
		t.Fatalf("create directory: %v", err)
	}

	idol := models.JavIdol{Name: "Preview Idol"}
	if err := db.Create(&idol).Error; err != nil {
		t.Fatalf("create idol: %v", err)
	}

	soloJav := models.Jav{Code: "DDD-001", Title: "Solo Work", Provider: 1, FetchedAt: now}
	groupJav := models.Jav{Code: "EEE-001", Title: "Group Work", Provider: 1, FetchedAt: now}
	coIdol := models.JavIdol{Name: "Other Idol"}
	if err := db.Create(&soloJav).Error; err != nil {
		t.Fatalf("create solo jav: %v", err)
	}
	if err := db.Create(&groupJav).Error; err != nil {
		t.Fatalf("create group jav: %v", err)
	}
	if err := db.Create(&coIdol).Error; err != nil {
		t.Fatalf("create co idol: %v", err)
	}

	maps := []models.JavIdolMap{
		{JavID: soloJav.ID, JavIdolID: idol.ID},
		{JavID: groupJav.ID, JavIdolID: idol.ID},
		{JavID: groupJav.ID, JavIdolID: coIdol.ID},
	}
	if err := db.Create(&maps).Error; err != nil {
		t.Fatalf("create idol maps: %v", err)
	}

	videos := []models.Video{
		{
			DirectoryID: dir.ID,
			Path:        "solo-preview.mp4",
			Filename:    "solo-preview.mp4",
			Fingerprint: "fp-solo-preview",
			JavID:       int64Ptr(soloJav.ID),
			ModifiedAt:  now,
		},
		{
			DirectoryID: dir.ID,
			Path:        "group-preview.mp4",
			Filename:    "group-preview.mp4",
			Fingerprint: "fp-group-preview",
			JavID:       int64Ptr(groupJav.ID),
			ModifiedAt:  now,
		},
	}
	if err := db.Create(&videos).Error; err != nil {
		t.Fatalf("create videos: %v", err)
	}

	item, err := GetJavIdolSummary(ctx, idol.ID)
	if err != nil {
		t.Fatalf("GetJavIdolSummary: %v", err)
	}
	if item.WorkCount != 2 {
		t.Fatalf("unexpected work count: got %d want 2", item.WorkCount)
	}
	if item.SampleCode != soloJav.Code {
		t.Fatalf("unexpected sample code: got %q want %q", item.SampleCode, soloJav.Code)
	}
}

func TestSearchJavSortByDurationDesc(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	now := time.Unix(1710000000, 0).UTC()

	dir := models.Directory{Path: "/tmp/media"}
	if err := db.Create(&dir).Error; err != nil {
		t.Fatalf("create directory: %v", err)
	}

	shortJav := models.Jav{
		Code:        "FFF-001",
		Title:       "Short",
		DurationMin: 90,
		Provider:    1,
		FetchedAt:   now,
	}
	longJav := models.Jav{
		Code:        "GGG-001",
		Title:       "Long",
		DurationMin: 180,
		Provider:    1,
		FetchedAt:   now,
	}
	if err := db.Create(&shortJav).Error; err != nil {
		t.Fatalf("create short jav: %v", err)
	}
	if err := db.Create(&longJav).Error; err != nil {
		t.Fatalf("create long jav: %v", err)
	}

	videos := []models.Video{
		{
			DirectoryID: dir.ID,
			Path:        "short.mp4",
			Filename:    "short.mp4",
			Fingerprint: "fp-short",
			JavID:       int64Ptr(shortJav.ID),
			ModifiedAt:  now,
		},
		{
			DirectoryID: dir.ID,
			Path:        "long.mp4",
			Filename:    "long.mp4",
			Fingerprint: "fp-long",
			JavID:       int64Ptr(longJav.ID),
			ModifiedAt:  now,
		},
	}
	if err := db.Create(&videos).Error; err != nil {
		t.Fatalf("create videos: %v", err)
	}

	items, total, err := SearchJav(ctx, nil, nil, "", "duration", 20, 0, nil)
	if err != nil {
		t.Fatalf("SearchJav: %v", err)
	}
	if total != 2 {
		t.Fatalf("unexpected total: got %d want 2", total)
	}
	if len(items) != 2 {
		t.Fatalf("unexpected item count: got %d want 2", len(items))
	}
	if items[0].ID != longJav.ID {
		t.Fatalf("unexpected first jav: got %d want %d", items[0].ID, longJav.ID)
	}
	if items[1].ID != shortJav.ID {
		t.Fatalf("unexpected second jav: got %d want %d", items[1].ID, shortJav.ID)
	}

	items, total, err = SearchJav(ctx, nil, nil, "", "duration_asc", 20, 0, nil)
	if err != nil {
		t.Fatalf("SearchJav duration_asc: %v", err)
	}
	if total != 2 {
		t.Fatalf("unexpected asc total: got %d want 2", total)
	}
	if len(items) != 2 {
		t.Fatalf("unexpected asc item count: got %d want 2", len(items))
	}
	if items[0].ID != shortJav.ID {
		t.Fatalf("unexpected asc first jav: got %d want %d", items[0].ID, shortJav.ID)
	}
	if items[1].ID != longJav.ID {
		t.Fatalf("unexpected asc second jav: got %d want %d", items[1].ID, longJav.ID)
	}
}

func TestListJavIdolsSortByAgeDirections(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	now := time.Unix(1710000000, 0).UTC()
	oldBirth := time.Date(1988, 1, 1, 0, 0, 0, 0, time.UTC)
	youngBirth := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)

	dir := models.Directory{Path: "/tmp/media"}
	if err := db.Create(&dir).Error; err != nil {
		t.Fatalf("create directory: %v", err)
	}

	oldIdol := models.JavIdol{Name: "Old Idol", BirthDate: &oldBirth}
	youngIdol := models.JavIdol{Name: "Young Idol", BirthDate: &youngBirth}
	if err := db.Create(&oldIdol).Error; err != nil {
		t.Fatalf("create old idol: %v", err)
	}
	if err := db.Create(&youngIdol).Error; err != nil {
		t.Fatalf("create young idol: %v", err)
	}

	oldJav := models.Jav{Code: "HHH-001", Title: "Old Solo", Provider: 1, FetchedAt: now}
	youngJav := models.Jav{Code: "III-001", Title: "Young Solo", Provider: 1, FetchedAt: now}
	if err := db.Create(&oldJav).Error; err != nil {
		t.Fatalf("create old jav: %v", err)
	}
	if err := db.Create(&youngJav).Error; err != nil {
		t.Fatalf("create young jav: %v", err)
	}

	maps := []models.JavIdolMap{
		{JavID: oldJav.ID, JavIdolID: oldIdol.ID},
		{JavID: youngJav.ID, JavIdolID: youngIdol.ID},
	}
	if err := db.Create(&maps).Error; err != nil {
		t.Fatalf("create idol maps: %v", err)
	}

	videos := []models.Video{
		{
			DirectoryID: dir.ID,
			Path:        "old.mp4",
			Filename:    "old.mp4",
			Fingerprint: "fp-old",
			JavID:       int64Ptr(oldJav.ID),
			ModifiedAt:  now,
		},
		{
			DirectoryID: dir.ID,
			Path:        "young.mp4",
			Filename:    "young.mp4",
			Fingerprint: "fp-young",
			JavID:       int64Ptr(youngJav.ID),
			ModifiedAt:  now,
		},
	}
	if err := db.Create(&videos).Error; err != nil {
		t.Fatalf("create videos: %v", err)
	}

	items, total, err := ListJavIdols(ctx, "", "birth", 20, 0)
	if err != nil {
		t.Fatalf("ListJavIdols birth: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Fatalf("unexpected birth result size: len=%d total=%d", len(items), total)
	}
	if items[0].ID != youngIdol.ID {
		t.Fatalf("unexpected birth first idol: got %d want %d", items[0].ID, youngIdol.ID)
	}

	items, total, err = ListJavIdols(ctx, "", "birth_asc", 20, 0)
	if err != nil {
		t.Fatalf("ListJavIdols birth_asc: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Fatalf("unexpected birth_asc result size: len=%d total=%d", len(items), total)
	}
	if items[0].ID != oldIdol.ID {
		t.Fatalf("unexpected birth_asc first idol: got %d want %d", items[0].ID, oldIdol.ID)
	}
}

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	prevDB := common.DB
	common.DB = db
	t.Cleanup(func() {
		common.DB = prevDB
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	return db
}

func int64Ptr(v int64) *int64 {
	return &v
}
