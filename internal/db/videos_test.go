package db

import (
	"context"
	"testing"
	"time"

	"pornboss/internal/models"
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

	items, err := ListVideos(ctx, 20, 0, nil, "", "duration", nil)
	if err != nil {
		t.Fatalf("ListVideos duration: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("unexpected item count: got %d want 2", len(items))
	}
	if items[0].ID != longVideo.ID {
		t.Fatalf("unexpected first video: got %d want %d", items[0].ID, longVideo.ID)
	}

	items, err = ListVideos(ctx, 20, 0, nil, "", "duration_asc", nil)
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

	items, err := ListVideos(ctx, 20, 0, nil, "", "recent", nil)
	if err != nil {
		t.Fatalf("ListVideos: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("unexpected item count: got %d want 2", len(items))
	}
	if len(items[0].Locations) != 1 || len(items[1].Locations) != 1 {
		t.Fatalf("list rows should be location-level: %#v", items)
	}
	count, err := CountVideos(ctx, nil, "")
	if err != nil {
		t.Fatalf("CountVideos: %v", err)
	}
	if count != 2 {
		t.Fatalf("unexpected location count: got %d want 2", count)
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
