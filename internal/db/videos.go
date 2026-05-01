package db

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"pornboss/internal/common"
	"pornboss/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ListVideos returns paginated video metadata ordered by the requested sort mode, filtered by all tagNames when provided.
func ListVideos(ctx context.Context, limit, offset int, tagNames []string, search, sort string, seed *int64) ([]models.Video, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	orderClause := "video.created_at DESC, video.id DESC" // default: newest first
	var orderExpr clause.Expr
	useExpr := false
	switch strings.ToLower(strings.TrimSpace(sort)) {
	case "filename", "filename_asc":
		orderClause = "video.filename COLLATE NOCASE, video.path"
	case "filename_desc":
		orderClause = "video.filename COLLATE NOCASE DESC, video.path DESC"
	case "duration", "duration_desc":
		orderClause = "video.duration_sec DESC, video.created_at DESC, video.id DESC"
	case "duration_asc":
		orderClause = "video.duration_sec ASC, video.created_at ASC, video.id ASC"
	case "play_count", "play_count_desc":
		orderClause = "COALESCE(video.play_count, 0) DESC, video.created_at DESC, video.id DESC"
	case "play_count_asc":
		orderClause = "COALESCE(video.play_count, 0) ASC, video.created_at ASC, video.id ASC"
	case "recent_asc":
		orderClause = "video.created_at ASC, video.id ASC"
	case "random":
		if seed != nil && *seed > 0 {
			orderExpr = clause.Expr{
				SQL:  "splitmix64(video.id, ?), video.id",
				Vars: []any{*seed},
			}
			useExpr = true
		} else {
			orderClause = "RANDOM()"
		}
	case "recent", "":
		// keep default
	default:
		// unknown value fallback to default
		orderClause = "video.created_at DESC, video.id DESC"
	}

	query := common.DB.WithContext(ctx).
		Model(&models.Video{}).
		Where("COALESCE(hidden, 0) = 0").
		Where("jav_id IS NULL").
		Preload("DirectoryRef").
		Limit(limit).
		Offset(offset)
	if useExpr {
		query = query.Order(clause.OrderBy{Expression: orderExpr})
	} else {
		query = query.Order(orderClause)
	}

	cleanedSearch := strings.TrimSpace(search)
	if cleanedSearch != "" {
		like := fmt.Sprintf("%%%s%%", cleanedSearch)
		query = query.Where("filename LIKE ? COLLATE NOCASE", like)
	}

	cleanedTags := normalizeTagNames(tagNames)
	if len(cleanedTags) > 0 {
		query = query.
			Joins("JOIN video_tag ON video_tag.video_id = video.id").
			Joins("JOIN tag ON tag.id = video_tag.tag_id").
			Where("tag.name IN ?", cleanedTags).
			Group("video.id").
			Having("COUNT(DISTINCT tag.name) = ?", len(cleanedTags))
	}

	var videos []models.Video
	if err := query.Preload("Tags").Find(&videos).Error; err != nil {
		return nil, fmt.Errorf("list videos: %w", err)
	}
	return videos, nil
}

// CountVideos returns the total number of videos that match optional tag filters + search term.
func CountVideos(ctx context.Context, tagNames []string, search string) (int64, error) {
	cleanedTags := normalizeTagNames(tagNames)
	cleanedSearch := strings.TrimSpace(search)
	like := ""
	if cleanedSearch != "" {
		like = fmt.Sprintf("%%%s%%", cleanedSearch)
	}

	// Base query counts videos; when filtering by tags, we need to group by video id and
	// ensure all requested tags are matched (intersection semantics).
	if len(cleanedTags) == 0 {
		base := common.DB.WithContext(ctx).
			Model(&models.Video{}).
			Where("COALESCE(hidden, 0) = 0").
			Where("jav_id IS NULL")
		if like != "" {
			base = base.Where("filename LIKE ? COLLATE NOCASE", like)
		}
		var count int64
		if err := base.Count(&count).Error; err != nil {
			return 0, fmt.Errorf("count videos: %w", err)
		}
		return count, nil
	}

	// Build subquery selecting matching video ids then count outer rows.
	sub := common.DB.WithContext(ctx).
		Model(&models.Video{}).
		Where("COALESCE(hidden, 0) = 0").
		Where("jav_id IS NULL").
		Select("video.id").
		Joins("JOIN video_tag ON video_tag.video_id = video.id").
		Joins("JOIN tag ON tag.id = video_tag.tag_id").
		Where("tag.name IN ?", cleanedTags)

	if like != "" {
		sub = sub.Where("filename LIKE ? COLLATE NOCASE", like)
	}

	sub = sub.Group("video.id").
		Having("COUNT(DISTINCT tag.name) = ?", len(cleanedTags))

	var count int64
	if err := common.DB.Table("(?) as m", sub).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count videos (filtered): %w", err)
	}
	return count, nil
}

// GetVideo fetches a single video by identifier.
func GetVideo(ctx context.Context, id int64) (*models.Video, error) {
	var video models.Video
	if err := common.DB.WithContext(ctx).
		Model(&models.Video{}).
		Where("COALESCE(hidden, 0) = 0").
		Preload("Tags").
		Preload("DirectoryRef").
		First(&video, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get video %d: %w", id, err)
	}
	return &video, nil
}

// AllVideos returns every video row; used for sync bookkeeping.
func AllVideos(ctx context.Context) ([]models.Video, error) {
	var videos []models.Video
	if err := common.DB.WithContext(ctx).Find(&videos).Error; err != nil {
		return nil, fmt.Errorf("load videos: %w", err)
	}
	return videos, nil
}

// ListUnhiddenVideoIDs returns IDs of videos that are not hidden.
func ListUnhiddenVideoIDs(ctx context.Context) ([]int64, error) {
	var ids []int64
	if err := common.DB.WithContext(ctx).
		Model(&models.Video{}).
		Where("COALESCE(hidden, 0) = 0").
		Pluck("id", &ids).Error; err != nil {
		return nil, fmt.Errorf("list unhidden video ids: %w", err)
	}
	return ids, nil
}

// SaveVideo inserts or updates a video based on its primary key.
func SaveVideo(ctx context.Context, video *models.Video) error {
	if video == nil {
		return errors.New("video is nil")
	}
	if err := common.DB.WithContext(ctx).Save(video).Error; err != nil {
		return fmt.Errorf("save video %q: %w", video.Path, err)
	}
	return nil
}

// CreateVideo inserts a new video record.
func CreateVideo(ctx context.Context, video *models.Video) error {
	if video == nil {
		return errors.New("video is nil")
	}
	if err := common.DB.WithContext(ctx).Create(video).Error; err != nil {
		return fmt.Errorf("create video %q: %w", video.Path, err)
	}
	return nil
}

// DeleteByIDs removes videos by their identifiers.
func DeleteByIDs(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	if err := common.DB.WithContext(ctx).Where("id IN ?", ids).Delete(&models.Video{}).Error; err != nil {
		return fmt.Errorf("delete videos: %w", err)
	}
	return nil
}

// HideVideosByIDs marks videos as hidden instead of deleting, preserving tags/JAV关联。
func HideVideosByIDs(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	if err := common.DB.WithContext(ctx).
		Model(&models.Video{}).
		Where("id IN ?", ids).
		Update("hidden", true).Error; err != nil {
		return fmt.Errorf("hide videos: %w", err)
	}
	return nil
}

// IncrementVideoPlayCount increments the play count for a video if it exists and is not hidden.
func IncrementVideoPlayCount(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("video id cannot be zero")
	}
	if err := common.DB.WithContext(ctx).
		Model(&models.Video{}).
		Where("id = ?", id).
		Where("COALESCE(hidden, 0) = 0").
		UpdateColumn("play_count", gorm.Expr("COALESCE(play_count, 0) + 1")).Error; err != nil {
		return fmt.Errorf("increment play count: %w", err)
	}
	return nil
}

// IncrementVideoPlayCountByPath increments the play count for a video located at a directory path + relative path.
func IncrementVideoPlayCountByPath(ctx context.Context, dirPath, relPath string) error {
	videoID, err := GetVideoIDByPath(ctx, dirPath, relPath)
	if err != nil {
		return err
	}
	if videoID == 0 {
		return nil
	}

	return IncrementVideoPlayCount(ctx, videoID)
}

// GetVideoIDByPath returns the ID for a video located at a directory path + relative path.
func GetVideoIDByPath(ctx context.Context, dirPath, relPath string) (int64, error) {
	if strings.TrimSpace(dirPath) == "" || strings.TrimSpace(relPath) == "" {
		return 0, errors.New("directory path and relative path are required")
	}

	var video models.Video
	err := common.DB.WithContext(ctx).
		Model(&models.Video{}).
		Select("video.id").
		Joins("JOIN directory ON directory.id = video.directory_id").
		Where("directory.path = ?", dirPath).
		Where("video.path = ?", relPath).
		Where("COALESCE(video.hidden, 0) = 0").
		First(&video).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, fmt.Errorf("lookup video by path: %w", err)
	}

	return video.ID, nil
}
