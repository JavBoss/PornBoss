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
	unavailableSoloJav := models.Jav{Code: "CCC-001", Title: "Unavailable Solo Work", Provider: 1, FetchedAt: now}
	if err := db.Create(&soloJav).Error; err != nil {
		t.Fatalf("create solo jav: %v", err)
	}
	if err := db.Create(&groupJav).Error; err != nil {
		t.Fatalf("create group jav: %v", err)
	}
	if err := db.Create(&unavailableSoloJav).Error; err != nil {
		t.Fatalf("create unavailable solo jav: %v", err)
	}

	maps := []models.JavIdolMap{
		{JavID: soloJav.ID, JavIdolID: soloIdol.ID},
		{JavID: groupJav.ID, JavIdolID: soloIdol.ID},
		{JavID: groupJav.ID, JavIdolID: groupOnlyIdol.ID},
		{JavID: unavailableSoloJav.ID, JavIdolID: groupOnlyIdol.ID},
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
			Path:        "unavailable.mp4",
			Filename:    "unavailable.mp4",
			Fingerprint: "fp-unavailable",
			JavID:       int64Ptr(unavailableSoloJav.ID),
			ModifiedAt:  now,
		},
	}
	if err := db.Create(&videos).Error; err != nil {
		t.Fatalf("create videos: %v", err)
	}
	if err := backfillVideoLocations(db); err != nil {
		t.Fatalf("backfill video locations: %v", err)
	}
	if err := db.Model(&models.VideoLocation{}).
		Where("video_id = ?", videos[2].ID).
		Update("is_delete", true).Error; err != nil {
		t.Fatalf("mark unavailable video location deleted: %v", err)
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

func TestJavBindingUsesVideoLocationsAndCountsLocations(t *testing.T) {
	gdb := openTestDB(t)
	ctx := context.Background()
	now := time.Unix(1710000000, 0).UTC()

	dir := models.Directory{Path: "/tmp/media"}
	if err := gdb.Create(&dir).Error; err != nil {
		t.Fatalf("create directory: %v", err)
	}
	video := models.Video{
		DirectoryID: dir.ID,
		Path:        "aaa-001.mp4",
		Filename:    "aaa-001.mp4",
		Fingerprint: "same-content-location-jav",
		DurationSec: 7200,
		ModifiedAt:  now,
	}
	if err := gdb.Create(&video).Error; err != nil {
		t.Fatalf("create video: %v", err)
	}

	javA := models.Jav{Code: "AAA-001", Title: "A", Provider: 1, FetchedAt: now}
	javB := models.Jav{Code: "BBB-001", Title: "B", Provider: 1, FetchedAt: now}
	if err := gdb.Create(&javA).Error; err != nil {
		t.Fatalf("create jav a: %v", err)
	}
	if err := gdb.Create(&javB).Error; err != nil {
		t.Fatalf("create jav b: %v", err)
	}
	tag := models.JavTag{Name: "Location Count", Provider: 1}
	if err := gdb.Create(&tag).Error; err != nil {
		t.Fatalf("create jav tag: %v", err)
	}
	idol := models.JavIdol{Name: "Location Idol"}
	if err := gdb.Create(&idol).Error; err != nil {
		t.Fatalf("create idol: %v", err)
	}
	if err := gdb.Create(&[]models.JavTagMap{{JavID: javA.ID, JavTagID: tag.ID}}).Error; err != nil {
		t.Fatalf("create tag map: %v", err)
	}
	if err := gdb.Create(&[]models.JavIdolMap{
		{JavID: javA.ID, JavIdolID: idol.ID},
		{JavID: javB.ID, JavIdolID: idol.ID},
	}).Error; err != nil {
		t.Fatalf("create idol maps: %v", err)
	}

	locs := []models.VideoLocation{
		{VideoID: video.ID, DirectoryID: dir.ID, RelativePath: "aaa-001-a.mp4", ModifiedAt: now, JavID: int64Ptr(javA.ID)},
		{VideoID: video.ID, DirectoryID: dir.ID, RelativePath: "aaa-001-b.mp4", ModifiedAt: now.Add(time.Second), JavID: int64Ptr(javA.ID)},
		{VideoID: video.ID, DirectoryID: dir.ID, RelativePath: "bbb-001.mp4", ModifiedAt: now.Add(2 * time.Second), JavID: int64Ptr(javB.ID)},
	}
	if err := gdb.Create(&locs).Error; err != nil {
		t.Fatalf("create locations: %v", err)
	}

	items, total, err := SearchJav(ctx, nil, nil, "", "code", 20, 0, nil)
	if err != nil {
		t.Fatalf("SearchJav: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Fatalf("unexpected jav result size: len=%d total=%d", len(items), total)
	}
	byCode := map[string]models.Jav{}
	for _, item := range items {
		byCode[item.Code] = item
	}
	if got := len(byCode["AAA-001"].Videos); got != 2 {
		t.Fatalf("AAA-001 video locations = %d, want 2", got)
	}
	if got := len(byCode["BBB-001"].Videos); got != 1 {
		t.Fatalf("BBB-001 video locations = %d, want 1", got)
	}
	if byCode["AAA-001"].Videos[0].ID != video.ID || byCode["BBB-001"].Videos[0].ID != video.ID {
		t.Fatal("expected location-backed videos to keep the original video id")
	}

	tags, err := ListJavTags(ctx)
	if err != nil {
		t.Fatalf("ListJavTags: %v", err)
	}
	tagCounts := map[string]int64{}
	for _, item := range tags {
		tagCounts[item.Name] = item.Count
	}
	if tagCounts[tag.Name] != 2 {
		t.Fatalf("tag count = %d, want 2", tagCounts[tag.Name])
	}

	idols, _, err := ListJavIdols(ctx, "", "work", 20, 0)
	if err != nil {
		t.Fatalf("ListJavIdols: %v", err)
	}
	if len(idols) != 1 || idols[0].ID != idol.ID {
		t.Fatalf("unexpected idols: %#v", idols)
	}
	if idols[0].WorkCount != 3 {
		t.Fatalf("idol work count = %d, want 3", idols[0].WorkCount)
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
	if err := backfillVideoLocations(db); err != nil {
		t.Fatalf("backfill video locations: %v", err)
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
	if err := backfillVideoLocations(db); err != nil {
		t.Fatalf("backfill video locations: %v", err)
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
	if err := backfillVideoLocations(db); err != nil {
		t.Fatalf("backfill video locations: %v", err)
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
