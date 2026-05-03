package db

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"pornboss/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestListVideosSortByDurationDirections(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	now := time.Unix(1710000000, 0).UTC()

	dir := models.Directory{Path: "/tmp/media"}
	if err := db.Create(&dir).Error; err != nil {
		t.Fatalf("create directory: %v", err)
	}

	shortVideo := models.Video{
		DirectoryID: dir.ID,
		Path:        "short.mp4",
		Filename:    "short.mp4",
		Fingerprint: "video-fp-short",
		DurationSec: 90,
		ModifiedAt:  now,
		CreatedAt:   now,
	}
	longVideo := models.Video{
		DirectoryID: dir.ID,
		Path:        "long.mp4",
		Filename:    "long.mp4",
		Fingerprint: "video-fp-long",
		DurationSec: 180,
		ModifiedAt:  now,
		CreatedAt:   now.Add(time.Second),
	}
	if err := db.Create(&shortVideo).Error; err != nil {
		t.Fatalf("create short video: %v", err)
	}
	if err := db.Create(&longVideo).Error; err != nil {
		t.Fatalf("create long video: %v", err)
	}
	if err := backfillVideoLocations(db); err != nil {
		t.Fatalf("backfill video locations: %v", err)
	}

	items, err := ListVideos(ctx, 20, 0, nil, "", "duration", nil, nil)
	if err != nil {
		t.Fatalf("ListVideos duration: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("unexpected item count: got %d want 2", len(items))
	}
	if items[0].ID != longVideo.ID {
		t.Fatalf("unexpected first video: got %d want %d", items[0].ID, longVideo.ID)
	}

	items, err = ListVideos(ctx, 20, 0, nil, "", "duration_asc", nil, nil)
	if err != nil {
		t.Fatalf("ListVideos duration_asc: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("unexpected asc item count: got %d want 2", len(items))
	}
	if items[0].ID != shortVideo.ID {
		t.Fatalf("unexpected asc first video: got %d want %d", items[0].ID, shortVideo.ID)
	}
}

func TestVideoLocationsAllowSameVideoInMultipleDirectories(t *testing.T) {
	gdb := openTestDB(t)
	ctx := context.Background()
	now := time.Unix(1710000000, 0).UTC()

	dirA := models.Directory{Path: "/tmp/media-a"}
	dirB := models.Directory{Path: "/tmp/media-b"}
	if err := gdb.Create(&dirA).Error; err != nil {
		t.Fatalf("create dir a: %v", err)
	}
	if err := gdb.Create(&dirB).Error; err != nil {
		t.Fatalf("create dir b: %v", err)
	}

	video := models.Video{
		DirectoryID: dirA.ID,
		Path:        "movie.mp4",
		Filename:    "movie.mp4",
		Fingerprint: "same-content",
		Size:        1024,
		DurationSec: 120,
		ModifiedAt:  now,
		Hidden:      true,
	}
	if err := gdb.Create(&video).Error; err != nil {
		t.Fatalf("create video: %v", err)
	}

	locA, err := UpsertVideoLocation(ctx, video.ID, dirA.ID, "movie.mp4", now)
	if err != nil {
		t.Fatalf("upsert loc a: %v", err)
	}
	locB, err := UpsertVideoLocation(ctx, video.ID, dirB.ID, "copies/movie.mp4", now.Add(time.Minute))
	if err != nil {
		t.Fatalf("upsert loc b: %v", err)
	}
	if err := ReconcileAllVideoPaths(ctx); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	items, err := ListVideos(ctx, 20, 0, nil, "", "recent", nil, nil)
	if err != nil {
		t.Fatalf("ListVideos: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("unexpected item count: got %d want 2", len(items))
	}
	if len(items[0].Locations) != 1 || len(items[1].Locations) != 1 {
		t.Fatalf("list rows should be location-level: %#v", items)
	}
	count, err := CountVideos(ctx, nil, "", nil)
	if err != nil {
		t.Fatalf("CountVideos: %v", err)
	}
	if count != 2 {
		t.Fatalf("unexpected location count: got %d want 2", count)
	}
	dirBItems, err := ListVideos(ctx, 20, 0, nil, "", "recent", nil, []int64{dirB.ID})
	if err != nil {
		t.Fatalf("ListVideos dir filter: %v", err)
	}
	if len(dirBItems) != 1 || dirBItems[0].LocationID != locB.ID || dirBItems[0].Path != "copies/movie.mp4" {
		t.Fatalf("unexpected filtered items: %#v", dirBItems)
	}
	dirBCount, err := CountVideos(ctx, nil, "", []int64{dirB.ID})
	if err != nil {
		t.Fatalf("CountVideos dir filter: %v", err)
	}
	if dirBCount != 1 {
		t.Fatalf("unexpected filtered count: got %d want 1", dirBCount)
	}
	pathSearchItems, err := ListVideos(ctx, 20, 0, nil, "copies", "recent", nil, nil)
	if err != nil {
		t.Fatalf("ListVideos path search: %v", err)
	}
	if len(pathSearchItems) != 0 {
		t.Fatalf("search should use filename, not relative path directories: %#v", pathSearchItems)
	}
	filenameSearchItems, err := ListVideos(ctx, 20, 0, nil, "movie.mp4", "recent", nil, nil)
	if err != nil {
		t.Fatalf("ListVideos filename search: %v", err)
	}
	if len(filenameSearchItems) != 2 {
		t.Fatalf("filename search should match both copies: got %d want 2", len(filenameSearchItems))
	}
	disabledItems, err := ListVideos(ctx, 20, 0, nil, "", "recent", nil, []int64{0})
	if err != nil {
		t.Fatalf("ListVideos disabled dir filter: %v", err)
	}
	if len(disabledItems) != 0 {
		t.Fatalf("disabled directory filter should return no rows: %#v", disabledItems)
	}
	disabledCount, err := CountVideos(ctx, nil, "", []int64{0})
	if err != nil {
		t.Fatalf("CountVideos disabled dir filter: %v", err)
	}
	if disabledCount != 0 {
		t.Fatalf("disabled directory filter count: got %d want 0", disabledCount)
	}

	videoID, err := GetVideoIDByPath(ctx, dirB.Path, "copies/movie.mp4")
	if err != nil {
		t.Fatalf("GetVideoIDByPath: %v", err)
	}
	if videoID != video.ID {
		t.Fatalf("unexpected video id by second location: got %d want %d", videoID, video.ID)
	}

	if err := HideVideoLocationsByIDs(ctx, []int64{locA.ID}); err != nil {
		t.Fatalf("hide loc a: %v", err)
	}
	if err := ReconcileAllVideoPaths(ctx); err != nil {
		t.Fatalf("reconcile after hiding loc a: %v", err)
	}
	visible, err := GetVideo(ctx, video.ID)
	if err != nil {
		t.Fatalf("GetVideo: %v", err)
	}
	if visible == nil {
		t.Fatal("video should remain visible while one location is active")
	}
	if len(visible.Locations) != 1 || visible.Locations[0].ID != locB.ID {
		t.Fatalf("unexpected remaining locations: %#v", visible.Locations)
	}

	if err := HideVideoLocationsByIDs(ctx, []int64{locB.ID}); err != nil {
		t.Fatalf("hide loc b: %v", err)
	}
	if err := ReconcileAllVideoPaths(ctx); err != nil {
		t.Fatalf("reconcile after hiding loc b: %v", err)
	}
	unavailable, err := GetVideo(ctx, video.ID)
	if err != nil {
		t.Fatalf("GetVideo unavailable: %v", err)
	}
	if unavailable != nil {
		t.Fatal("video should be unavailable when all locations are deleted")
	}
}

func TestBackfillVideoLocationsOnceUsesMarker(t *testing.T) {
	gdb := openTestDB(t)
	now := time.Unix(1710000000, 0).UTC()

	if err := gdb.Where("key = ?", videoLocationBackfillMarkerKey).Delete(&models.Config{}).Error; err != nil {
		t.Fatalf("delete backfill marker: %v", err)
	}
	dir := models.Directory{Path: "/tmp/media"}
	if err := gdb.Create(&dir).Error; err != nil {
		t.Fatalf("create directory: %v", err)
	}
	javRec := models.Jav{Code: "AAA-001", Title: "A", Provider: 1, FetchedAt: now}
	if err := gdb.Create(&javRec).Error; err != nil {
		t.Fatalf("create jav: %v", err)
	}
	video := models.Video{
		DirectoryID: dir.ID,
		Path:        "aaa-001.mp4",
		Filename:    "aaa-001.mp4",
		Fingerprint: "legacy-fp",
		ModifiedAt:  now,
		JavID:       &javRec.ID,
	}
	if err := gdb.Create(&video).Error; err != nil {
		t.Fatalf("create legacy-style video: %v", err)
	}

	if err := backfillVideoLocationsOnce(gdb); err != nil {
		t.Fatalf("first backfill: %v", err)
	}
	var loc models.VideoLocation
	if err := gdb.Where("video_id = ?", video.ID).First(&loc).Error; err != nil {
		t.Fatalf("load backfilled location: %v", err)
	}
	if loc.JavID == nil || *loc.JavID != javRec.ID {
		t.Fatalf("backfilled jav_id = %v, want %d", loc.JavID, javRec.ID)
	}
	if err := gdb.Model(&models.VideoLocation{}).Where("id = ?", loc.ID).Update("jav_id", nil).Error; err != nil {
		t.Fatalf("clear location jav_id: %v", err)
	}

	if err := backfillVideoLocationsOnce(gdb); err != nil {
		t.Fatalf("second backfill: %v", err)
	}
	var after models.VideoLocation
	if err := gdb.Where("id = ?", loc.ID).First(&after).Error; err != nil {
		t.Fatalf("reload location: %v", err)
	}
	if after.JavID != nil {
		t.Fatalf("jav_id was backfilled again after marker: got %v", *after.JavID)
	}

	if err := gdb.Where("key = ?", videoLocationBackfillMarkerKey).Delete(&models.Config{}).Error; err != nil {
		t.Fatalf("delete marker after existing locations: %v", err)
	}
	if err := backfillVideoLocationsOnce(gdb); err != nil {
		t.Fatalf("backfill with existing locations and no marker: %v", err)
	}
	var existingAfter models.VideoLocation
	if err := gdb.Where("id = ?", loc.ID).First(&existingAfter).Error; err != nil {
		t.Fatalf("reload existing location: %v", err)
	}
	if existingAfter.JavID != nil {
		t.Fatalf("jav_id was backfilled again for existing locations: got %v", *existingAfter.JavID)
	}
}

type legacyVideoForMigration struct {
	ID          int64  `gorm:"primaryKey"`
	DirectoryID int64  `gorm:"index;not null"`
	Path        string `gorm:"index"`
	Filename    string
	Size        int64
	ModifiedAt  time.Time
	Fingerprint string `gorm:"uniqueIndex"`
	DurationSec int64
	PlayCount   int64 `gorm:"not null;default:0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	JavID       *int64           `gorm:"index"`
	Jav         *models.Jav      `gorm:"foreignKey:JavID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Directory   models.Directory `gorm:"foreignKey:DirectoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Hidden      bool             `gorm:"index"`
	Tags        []models.Tag     `gorm:"many2many:video_tag"`
}

func (legacyVideoForMigration) TableName() string {
	return "video"
}

func TestOpenMigratesLegacyGormSchemaPreservesVideoTags(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "legacy-gorm.db")
	legacy, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		t.Fatalf("open legacy db: %v", err)
	}
	if err := legacy.AutoMigrate(
		&models.Directory{},
		&legacyVideoForMigration{},
		&models.Tag{},
		&models.VideoTag{},
		&models.Config{},
		&models.Jav{},
		&models.JavTag{},
		&models.JavIdol{},
		&models.JavTagMap{},
		&models.JavIdolMap{},
	); err != nil {
		t.Fatalf("create legacy schema: %v", err)
	}

	now := time.Unix(1710000000, 0).UTC()
	dir := models.Directory{Path: "/tmp/media"}
	if err := legacy.Create(&dir).Error; err != nil {
		t.Fatalf("create legacy directory: %v", err)
	}
	video := legacyVideoForMigration{
		DirectoryID: dir.ID,
		Path:        "tagged.mp4",
		Filename:    "tagged.mp4",
		Fingerprint: "legacy-tagged-fp",
		ModifiedAt:  now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := legacy.Create(&video).Error; err != nil {
		t.Fatalf("create legacy video: %v", err)
	}
	tag := models.Tag{Name: "favorite", CreatedAt: now, UpdatedAt: now}
	if err := legacy.Create(&tag).Error; err != nil {
		t.Fatalf("create legacy tag: %v", err)
	}
	if err := legacy.Create(&models.VideoTag{VideoID: video.ID, TagID: tag.ID, CreatedAt: now}).Error; err != nil {
		t.Fatalf("create legacy video tag: %v", err)
	}
	var beforeCount int64
	if err := legacy.Model(&models.VideoTag{}).Count(&beforeCount).Error; err != nil {
		t.Fatalf("count legacy video tags: %v", err)
	}
	if beforeCount != 1 {
		t.Fatalf("legacy video_tag count = %d, want 1", beforeCount)
	}
	legacySQL, err := legacy.DB()
	if err != nil {
		t.Fatalf("legacy sql db: %v", err)
	}
	_ = legacySQL.Close()

	migrated, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open migrated db: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := migrated.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	var afterCount int64
	if err := migrated.Model(&models.VideoTag{}).Where("video_id = ? AND tag_id = ?", video.ID, tag.ID).Count(&afterCount).Error; err != nil {
		t.Fatalf("count migrated video tags: %v", err)
	}
	if afterCount != 1 {
		t.Fatalf("migrated video_tag count = %d, want 1", afterCount)
	}
}
