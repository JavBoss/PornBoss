package service

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"pornboss/internal/common"
	"pornboss/internal/common/logging"
	"pornboss/internal/db"
	"pornboss/internal/models"
	"pornboss/internal/util"
)

type FileEntry struct {
	DirectoryID   int64
	DirectoryPath string
	RelativePath  string
	AbsolutePath  string
	Filename      string
	Size          int64
	ModifiedAt    time.Time
	Fingerprint   string
	PathKey       string
	DurationSec   int64
}

// Summary captures the results of a directory synchronisation run.
type Summary struct {
	FilesSeen   int
	Inserted    int
	Updated     int
	Removed     int
	Duration    time.Duration
	Directories int
}

func makePathKey(directoryID int64, relativePath string) string {
	return strconv.FormatInt(directoryID, 10) + ":" + relativePath
}

// SyncAllDirectories scans multiple directories in a single pass so cross-directory moves keep tags/metadata.
func SyncAllDirectories(ctx context.Context, directories []models.Directory) (*Summary, error) {
	if common.DB == nil {
		return nil, errors.New("nil database")
	}
	if len(directories) == 0 {
		return &Summary{}, nil
	}

	locked := make([]int64, 0, len(directories))
	unlockAll := func() {
		for _, id := range locked {
			unlockDirectory(id)
		}
	}
	for _, dir := range directories {
		if dir.ID == 0 || dir.IsDelete {
			continue
		}
		if !tryLockDirectory(dir.ID) {
			unlockAll()
			return nil, ErrDirectoryScanInProgress
		}
		locked = append(locked, dir.ID)
	}
	defer unlockAll()

	start := time.Now()
	summary := &Summary{}

	existing, err := db.AllVideos(ctx)
	if err != nil {
		return nil, err
	}
	existingByFingerprint := make(map[string]*models.Video, len(existing))
	existingByID := make(map[int64]*models.Video, len(existing))
	for i := range existing {
		video := &existing[i]
		existingByID[video.ID] = video
		if video.Fingerprint != "" {
			existingByFingerprint[video.Fingerprint] = video
		}
	}
	existingLocations, err := db.AllVideoLocations(ctx)
	if err != nil {
		return nil, err
	}
	existingLocationByPath := make(map[string]*models.VideoLocation, len(existingLocations))
	processedLocationIDs := make(map[int64]struct{}, len(existingLocations))
	for i := range existingLocations {
		loc := &existingLocations[i]
		existingLocationByPath[makePathKey(loc.DirectoryID, loc.RelativePath)] = loc
	}

	scanned := make(map[int64]struct{}, len(directories))
	for _, dir := range directories {
		if dir.ID == 0 || dir.IsDelete {
			continue
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		info, statErr := os.Stat(dir.Path)
		if statErr != nil {
			if util.IsPathUnavailable(statErr) {
				if err := db.SetDirectoryMissingAndHideVideos(ctx, dir.ID, true); err != nil {
					logging.Error("mark directory missing failed id=%d path=%s err=%v", dir.ID, dir.Path, err)
				}
				continue
			}
			return nil, fmt.Errorf("stat directory %s: %w", dir.Path, statErr)
		}
		if !info.IsDir() {
			if err := db.SetDirectoryMissingAndHideVideos(ctx, dir.ID, true); err != nil {
				logging.Error("mark directory missing failed id=%d path=%s err=%v", dir.ID, dir.Path, err)
			}
			continue
		}
		if dir.Missing {
			if err := db.SetDirectoryMissingAndHideVideos(ctx, dir.ID, false); err != nil {
				logging.Error("clear directory missing failed id=%d path=%s err=%v", dir.ID, dir.Path, err)
			}
		}

		scanned[dir.ID] = struct{}{}
		if err := walkDirectoryAndSync(ctx, dir, existingByID, existingByFingerprint, existingLocationByPath, processedLocationIDs, summary); err != nil {
			return nil, err
		}
	}

	if err := hideUnprocessedVideoLocations(ctx, processedLocationIDs, summary); err != nil {
		return nil, err
	}

	summary.Directories = len(scanned)
	summary.Duration = time.Since(start)
	return summary, nil
}

// walkDirectoryAndSync 扫描单个目录下的文件，实时 upsert 并记录统计/缩略图任务。
func walkDirectoryAndSync(ctx context.Context, directory models.Directory, existingByID map[int64]*models.Video, existingByFingerprint map[string]*models.Video, existingLocationByPath map[string]*models.VideoLocation, processedLocationIDs map[int64]struct{}, summary *Summary) error {
	// 边遍历文件边做指纹计算和 DB 更新，避免一次性构建全量快照
	normalizedRoot := filepath.Clean(directory.Path)
	return filepath.WalkDir(normalizedRoot, func(candidatePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if errors.Is(walkErr, fs.ErrPermission) {
				return nil
			}
			return walkErr
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}
		if !util.IsVideo(candidatePath) {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			if errors.Is(err, fs.ErrPermission) {
				return nil
			}
			return err
		}

		// 计算相对路径，确保只处理目录内的文件（防止符号链接等越界）
		normalizedPath := filepath.Clean(candidatePath)
		relativePath, err := filepath.Rel(normalizedRoot, normalizedPath)
		if err != nil {
			return err
		}
		if strings.HasPrefix(relativePath, "..") {
			return nil
		}

		relativePath = filepath.ToSlash(cleanRelativePath(relativePath))
		modTime := info.ModTime().UTC()
		summary.FilesSeen++

		// If file unchanged (size + mtime), skip probe and DB touches but mark as seen.
		pathKey := makePathKey(directory.ID, relativePath)
		if existingLoc, ok := existingLocationByPath[pathKey]; ok {
			existingVideo := existingByID[existingLoc.VideoID]
			if existingVideo != nil && existingVideo.Size == info.Size() && existingLoc.ModifiedAt.Equal(modTime) {
				processedLocationIDs[existingLoc.ID] = struct{}{}
				if existingLoc.IsDelete {
					saved, err := db.UpsertVideoLocation(ctx, existingLoc.VideoID, directory.ID, relativePath, modTime)
					if err != nil {
						logging.Error("unhide video location failed, skip: path=%s err=%v", normalizedPath, err)
						return nil
					}
					existingLoc.IsDelete = false
					existingLoc.ModifiedAt = modTime
					if saved != nil {
						processedLocationIDs[saved.ID] = struct{}{}
					}
					summary.Updated++
				}
				return nil
			}
		}

		logging.Info("scan file: root=%s path=%s size=%d", normalizedRoot, relativePath, info.Size())

		meta, err := util.ProbeVideo(normalizedPath)
		if err != nil {
			logging.Error("probe video metadata error: %v", err)
			return nil
		}
		fingerprint := meta.FingerprintV2(info.Size())
		durationSec := int64(math.Round(meta.DurationSeconds))

		fileEntry := &FileEntry{
			DirectoryID:   directory.ID,
			DirectoryPath: normalizedRoot,
			RelativePath:  relativePath,
			AbsolutePath:  normalizedPath,
			Filename:      info.Name(),
			Size:          info.Size(),
			ModifiedAt:    modTime,
			Fingerprint:   fingerprint,
			DurationSec:   durationSec,
		}

		return upsertVideo(ctx, fileEntry, existingByID, existingByFingerprint, existingLocationByPath, processedLocationIDs, summary)
	})
}

func upsertVideo(ctx context.Context, entry *FileEntry, existingByID map[int64]*models.Video, existingByFingerprint map[string]*models.Video, existingLocationByPath map[string]*models.VideoLocation, processedLocationIDs map[int64]struct{}, summary *Summary) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// 已存在：检测元信息/文件是否变化，必要时更新
	var video *models.Video
	if entry.Fingerprint != "" {
		if existingVideo, ok := existingByFingerprint[entry.Fingerprint]; ok {
			video = existingVideo
		}
	}

	if video != nil {
		if video.Size != entry.Size || video.DurationSec != entry.DurationSec || video.Fingerprint != entry.Fingerprint {
			video.Size = entry.Size
			video.Fingerprint = entry.Fingerprint
			video.DurationSec = entry.DurationSec

			if err := db.SaveVideo(ctx, video); err != nil {
				logging.Error("save video failed, skip: path=%s err=%v", entry.AbsolutePath, err)
				return nil
			}
			summary.Updated++
		}

		if err := upsertLocationForEntry(ctx, video, entry, existingByID, existingLocationByPath, processedLocationIDs); err != nil {
			logging.Error("save video location failed, skip: path=%s err=%v", entry.AbsolutePath, err)
			return nil
		}
		return nil
	}

	video = &models.Video{
		DirectoryID: entry.DirectoryID,
		Size:        entry.Size,
		Fingerprint: entry.Fingerprint,
		DurationSec: entry.DurationSec,
	}

	if err := db.CreateVideo(ctx, video); err != nil {
		logging.Error("create video failed, skip: path=%s err=%v", entry.AbsolutePath, err)
		return nil
	}
	summary.Inserted++
	existingByID[video.ID] = video
	if video.Fingerprint != "" {
		existingByFingerprint[video.Fingerprint] = video
	}
	if err := upsertLocationForEntry(ctx, video, entry, existingByID, existingLocationByPath, processedLocationIDs); err != nil {
		logging.Error("save video location failed, skip: path=%s err=%v", entry.AbsolutePath, err)
		return nil
	}
	video.ModifiedAt = entry.ModifiedAt
	common.ScreenshotManager.EnqueueForVideo(video)
	return nil
}

func upsertLocationForEntry(ctx context.Context, video *models.Video, entry *FileEntry, existingByID map[int64]*models.Video, existingLocationByPath map[string]*models.VideoLocation, processedLocationIDs map[int64]struct{}) error {
	loc, err := db.UpsertVideoLocation(ctx, video.ID, entry.DirectoryID, entry.RelativePath, entry.ModifiedAt)
	if err != nil {
		return err
	}
	if loc != nil {
		processedLocationIDs[loc.ID] = struct{}{}
		existingLocationByPath[makePathKey(loc.DirectoryID, loc.RelativePath)] = loc
	}
	existingByID[video.ID] = video
	return nil
}

// hideUnprocessedVideoLocations marks file locations missing from this scan as deleted.
func hideUnprocessedVideoLocations(ctx context.Context, processedLocationIDs map[int64]struct{}, summary *Summary) error {
	locations, err := db.AllVideoLocations(ctx)
	if err != nil {
		return err
	}
	if len(locations) == 0 {
		return nil
	}

	staleIDs := make([]int64, 0, len(locations))
	for _, loc := range locations {
		if loc.IsDelete {
			continue
		}
		if _, ok := processedLocationIDs[loc.ID]; ok {
			continue
		}
		staleIDs = append(staleIDs, loc.ID)
	}

	if len(staleIDs) == 0 {
		return nil
	}

	logging.Info("hiding stale video locations: count=%d", len(staleIDs))
	if err := db.HideVideoLocationsByIDs(ctx, staleIDs); err != nil {
		return err
	}
	summary.Removed += len(staleIDs)
	return nil
}

func cleanRelativePath(p string) string {
	if p == "" {
		return ""
	}
	cleaned := filepath.Clean(p)
	if cleaned == "." {
		return ""
	}
	return filepath.ToSlash(cleaned)
}
