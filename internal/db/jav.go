package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"pornboss/internal/common"
	"pornboss/internal/jav"
	"pornboss/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// JavTagCount represents a JAV tag with associated work count.
type JavTagCount struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Provider int    `json:"provider"`
	Count    int64  `json:"count"`
}

// JavScanVideo contains the fields the scanner needs to resolve or refresh JAV metadata.
type JavScanVideo struct {
	ID          int64     `gorm:"column:id"`
	Filename    string    `gorm:"column:filename"`
	JavID       *int64    `gorm:"column:jav_id"`
	JavProvider int       `gorm:"column:jav_provider"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
	DurationSec int64     `gorm:"column:duration_sec"`
}

// SearchJav lists Jav metadata filtered by actors/tag IDs/search with pagination and sorting.
func SearchJav(ctx context.Context, actors []string, tagIDs []int64, search, sort string, limit, offset int, seed *int64) ([]models.Jav, int64, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	actors = normalizeNames(actors)
	tagIDs = uniqueInt64s(tagIDs)
	search = strings.TrimSpace(search)
	sort = strings.ToLower(strings.TrimSpace(sort))

	filtered := buildJavFilter(ctx, actors, tagIDs, search)

	// Count on a cloned query to avoid mutating the main one.
	countBase := buildJavFilter(ctx, actors, tagIDs, search)
	countQuery := countBase.Select("DISTINCT jav.id")
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count jav: %w", err)
	}

	var items []models.Jav
	order := "jav.created_at DESC"
	var orderExpr clause.Expr
	useExpr := false
	switch sort {
	case "code":
		order = "jav.code"
	case "release":
		order = "jav.release_unix DESC, jav.code"
	case "play_count":
		order = "COALESCE((SELECT SUM(COALESCE(v.play_count, 0)) FROM video v WHERE v.jav_id = jav.id AND COALESCE(v.hidden, 0) = 0), 0) DESC, jav.created_at DESC, jav.id DESC"
	case "random":
		if seed != nil && *seed > 0 {
			orderExpr = clause.Expr{
				SQL:  "splitmix64(jav.id, ?), jav.id",
				Vars: []any{*seed},
			}
			useExpr = true
		} else {
			order = "RANDOM()"
		}
	case "recent", "":
		// default order
	default:
		// ignore unknown values
	}
	visibleTagProviders := visibleJavTagProviders()
	query := filtered.
		Preload("Tags", "provider IN ?", visibleTagProviders).
		Preload("Idols").
		Preload("Videos", "COALESCE(hidden, 0) = 0").
		Preload("Videos.DirectoryRef").
		Limit(limit).
		Offset(offset)
	if useExpr {
		query = query.Order(clause.OrderBy{Expression: orderExpr})
	} else {
		query = query.Order(order)
	}
	if err := query.Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("list jav: %w", err)
	}
	return items, total, nil
}

// ListJavTags returns JAV tags with the number of works for each tag.
func ListJavTags(ctx context.Context) ([]JavTagCount, error) {
	var tags []JavTagCount
	visibleProviders := visibleJavTagProviders()
	if err := common.DB.WithContext(ctx).
		Table("jav_tag jt").
		Select("jt.id, jt.name, jt.provider, COUNT(DISTINCT CASE WHEN COALESCE(v.hidden, 0) = 0 THEN jtm.jav_id END) AS count").
		Joins("LEFT JOIN jav_tag_map jtm ON jtm.jav_tag_id = jt.id").
		Joins("LEFT JOIN video v ON v.jav_id = jtm.jav_id").
		Where("jt.provider IN ?", visibleProviders).
		Group("jt.id, jt.name, jt.provider").
		Order("jt.name, jt.provider").
		Scan(&tags).Error; err != nil {
		return nil, fmt.Errorf("list jav tags: %w", err)
	}
	return tags, nil
}

func visibleJavTagProviders() []int {
	current := int(jav.PreferredProvider())
	if current <= 0 {
		current = int(jav.ProviderJavBus)
	}
	return []int{current, int(jav.ProviderUser)}
}

// CreateJavTag creates a user-defined JAV tag.
func CreateJavTag(ctx context.Context, name string) (*models.JavTag, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("tag name cannot be empty")
	}

	tag := models.JavTag{Name: name, Provider: int(jav.ProviderUser)}
	if err := common.DB.WithContext(ctx).Create(&tag).Error; err != nil {
		return nil, fmt.Errorf("create jav tag %q: %w", name, err)
	}
	return &tag, nil
}

// RenameJavTag renames a user-created JAV tag.
func RenameJavTag(ctx context.Context, id int64, newName string) error {
	newName = strings.TrimSpace(newName)
	if id == 0 {
		return errors.New("tag id cannot be zero")
	}
	if newName == "" {
		return errors.New("tag name cannot be empty")
	}

	var tag models.JavTag
	if err := common.DB.WithContext(ctx).First(&tag, id).Error; err != nil {
		return fmt.Errorf("find jav tag: %w", err)
	}
	if !isUserJavTagProvider(tag.Provider) {
		return errors.New("tag is not user-defined")
	}

	if err := common.DB.WithContext(ctx).
		Model(&models.JavTag{}).
		Where("id = ? AND provider = ?", id, int(jav.ProviderUser)).
		Update("name", newName).Error; err != nil {
		return fmt.Errorf("rename jav tag: %w", err)
	}
	return nil
}

// DeleteJavTag removes a user-created JAV tag and detaches it from any associated entries.
func DeleteJavTag(ctx context.Context, id int64) error {
	if id == 0 {
		return errors.New("tag id cannot be zero")
	}

	var tag models.JavTag
	if err := common.DB.WithContext(ctx).First(&tag, id).Error; err != nil {
		return fmt.Errorf("find jav tag: %w", err)
	}
	if !isUserJavTagProvider(tag.Provider) {
		return errors.New("tag is not user-defined")
	}

	return common.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("jav_tag_id = ?", id).Delete(&models.JavTagMap{}).Error; err != nil {
			return fmt.Errorf("delete jav tag relations: %w", err)
		}
		if err := tx.Delete(&models.JavTag{}, id).Error; err != nil {
			return fmt.Errorf("delete jav tag: %w", err)
		}
		return nil
	})
}

// DeleteJavTags removes multiple user-created JAV tags and detaches them.
func DeleteJavTags(ctx context.Context, ids []int64) error {
	cleanIDs := uniqueInt64s(ids)
	if len(cleanIDs) == 0 {
		return nil
	}

	var count int64
	if err := common.DB.WithContext(ctx).
		Model(&models.JavTag{}).
		Where("id IN ? AND provider = ?", cleanIDs, int(jav.ProviderUser)).
		Count(&count).Error; err != nil {
		return fmt.Errorf("find jav tags: %w", err)
	}
	if count != int64(len(cleanIDs)) {
		return errors.New("tag is not user-defined")
	}

	return common.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("jav_tag_id IN ?", cleanIDs).Delete(&models.JavTagMap{}).Error; err != nil {
			return fmt.Errorf("delete jav tag relations: %w", err)
		}
		if err := tx.Where("id IN ?", cleanIDs).Delete(&models.JavTag{}).Error; err != nil {
			return fmt.Errorf("delete jav tag: %w", err)
		}
		return nil
	})
}

// AddJavTagToJavs associates a user-created tag with JAV entries.
func AddJavTagToJavs(ctx context.Context, tagID int64, javIDs []int64) error {
	if tagID == 0 || len(javIDs) == 0 {
		return nil
	}
	cleanIDs := uniqueInt64s(javIDs)
	if len(cleanIDs) == 0 {
		return nil
	}

	return common.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var tag models.JavTag
		if err := tx.First(&tag, tagID).Error; err != nil {
			return fmt.Errorf("find jav tag: %w", err)
		}
		if !isUserJavTagProvider(tag.Provider) {
			return errors.New("tag is not user-defined")
		}

		now := time.Now()
		rows := make([]models.JavTagMap, 0, len(cleanIDs))
		for _, javID := range cleanIDs {
			rows = append(rows, models.JavTagMap{JavID: javID, JavTagID: tagID, CreatedAt: now})
		}
		if len(rows) == 0 {
			return nil
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error; err != nil {
			return fmt.Errorf("insert jav tag map: %w", err)
		}
		return nil
	})
}

// RemoveJavTagFromJavs detaches a user-created tag from JAV entries.
func RemoveJavTagFromJavs(ctx context.Context, tagID int64, javIDs []int64) error {
	if tagID == 0 || len(javIDs) == 0 {
		return nil
	}
	cleanIDs := uniqueInt64s(javIDs)
	if len(cleanIDs) == 0 {
		return nil
	}

	var tag models.JavTag
	if err := common.DB.WithContext(ctx).First(&tag, tagID).Error; err != nil {
		return fmt.Errorf("find jav tag: %w", err)
	}
	if !isUserJavTagProvider(tag.Provider) {
		return errors.New("tag is not user-defined")
	}

	if err := common.DB.WithContext(ctx).
		Where("jav_id IN ? AND jav_tag_id = ?", cleanIDs, tagID).
		Delete(&models.JavTagMap{}).Error; err != nil {
		return fmt.Errorf("delete jav tag map: %w", err)
	}
	return nil
}

// ReplaceJavUserTags replaces user-defined tags for the given JAV entries.
func ReplaceJavUserTags(ctx context.Context, javIDs, tagIDs []int64) error {
	cleanJavIDs := uniqueInt64s(javIDs)
	if len(cleanJavIDs) == 0 {
		return nil
	}
	cleanTagIDs := uniqueInt64s(tagIDs)

	if len(cleanTagIDs) > 0 {
		var count int64
		if err := common.DB.WithContext(ctx).
			Model(&models.JavTag{}).
			Where("id IN ? AND provider = ?", cleanTagIDs, int(jav.ProviderUser)).
			Count(&count).Error; err != nil {
			return fmt.Errorf("find jav tags: %w", err)
		}
		if count != int64(len(cleanTagIDs)) {
			return errors.New("invalid tag_id")
		}
	}

	return common.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Where("jav_id IN ? AND jav_tag_id IN (SELECT id FROM jav_tag WHERE provider = ?)", cleanJavIDs, int(jav.ProviderUser)).
			Delete(&models.JavTagMap{}).Error; err != nil {
			return fmt.Errorf("delete jav tag map: %w", err)
		}
		if len(cleanTagIDs) == 0 {
			return nil
		}
		now := time.Now()
		rows := make([]models.JavTagMap, 0, len(cleanJavIDs)*len(cleanTagIDs))
		for _, javID := range cleanJavIDs {
			for _, tagID := range cleanTagIDs {
				rows = append(rows, models.JavTagMap{JavID: javID, JavTagID: tagID, CreatedAt: now})
			}
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error; err != nil {
			return fmt.Errorf("insert jav tag map: %w", err)
		}
		return nil
	})
}

func buildJavFilter(ctx context.Context, actors []string, tagIDs []int64, search string) *gorm.DB {
	q := common.DB.WithContext(ctx).Model(&models.Jav{})
	visibleTagProviders := visibleJavTagProviders()
	// Only include JAV entries that have at least one visible video.
	validVideo := common.DB.WithContext(ctx).
		Table("video v").
		Select("1").
		Where("v.jav_id = jav.id").
		Where("COALESCE(v.hidden, 0) = 0")
	q = q.Where("EXISTS (?)", validVideo)
	if search != "" {
		like := fmt.Sprintf("%%%s%%", search)
		q = q.Where("code LIKE ? OR title LIKE ?", like, like)
	}
	if len(tagIDs) > 0 {
		q = q.
			Joins("JOIN jav_tag_map jtm ON jtm.jav_id = jav.id").
			Joins("JOIN jav_tag jt ON jt.id = jtm.jav_tag_id").
			Where("jt.provider IN ?", visibleTagProviders).
			Where("jtm.jav_tag_id IN ?", tagIDs).
			Group("jav.id").
			Having("COUNT(DISTINCT jtm.jav_tag_id) = ?", len(tagIDs))
	}
	if len(actors) > 0 {
		q = q.
			Joins("JOIN jav_idol_map jim ON jim.jav_id = jav.id").
			Joins("JOIN jav_idol ji ON ji.id = jim.jav_idol_id").
			Where("ji.name IN ?", actors).
			Group("jav.id").
			Having("COUNT(DISTINCT ji.name) >= ?", len(actors))
	}
	return q
}

// JavIdolSummary represents idol info with aggregated work count and a sample code for cover lookup.
type JavIdolSummary struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name"`
	RomanName    string     `json:"roman_name"`
	JapaneseName string     `json:"japanese_name"`
	ChineseName  string     `json:"chinese_name"`
	HeightCM     *int       `json:"height_cm"`
	BirthDate    *time.Time `json:"birth_date"`
	Bust         *int       `json:"bust"`
	Waist        *int       `json:"waist"`
	Hips         *int       `json:"hips"`
	Cup          *int       `json:"cup"`
	WorkCount    int64      `json:"work_count"`
	SampleCode   string     `json:"sample_code"`
}

// ListJavIdols returns idols ordered by selected sort with pagination.
func ListJavIdols(ctx context.Context, search, sort string, limit, offset int) ([]JavIdolSummary, int64, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	search = strings.TrimSpace(search)
	sort = strings.ToLower(strings.TrimSpace(sort))

	base := common.DB.WithContext(ctx).
		Table("jav_idol ji").
		Joins("JOIN jav_idol_map jim ON jim.jav_idol_id = ji.id").
		Joins("JOIN jav j ON j.id = jim.jav_id").
		Joins("JOIN video v ON v.jav_id = j.id").
		Where("COALESCE(v.hidden, 0) = 0")

	if search != "" {
		like := fmt.Sprintf("%%%s%%", search)
		base = base.Where(
			"ji.name LIKE ? OR ji.japanese_name LIKE ? OR ji.roman_name LIKE ? OR ji.chinese_name LIKE ?",
			like,
			like,
			like,
			like,
		)
	}

	var total int64
	if err := base.Select("COUNT(DISTINCT ji.id)").Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count jav idols: %w", err)
	}

	var items []JavIdolSummary
	soloCode := `
COALESCE(
  (
    SELECT j2.code
    FROM jav_idol_map jim2
    JOIN jav j2 ON j2.id = jim2.jav_id
    JOIN video v2 ON v2.jav_id = j2.id
    WHERE jim2.jav_idol_id = ji.id
      AND COALESCE(v2.hidden, 0) = 0
      AND j2.id IN (SELECT jav_id FROM jav_idol_map GROUP BY jav_id HAVING COUNT(*) = 1)
    ORDER BY j2.code
    LIMIT 1
  ),
  MIN(j.code)
) AS sample_code`
	order := "work_count DESC, ji.name ASC"
	switch sort {
	case "birth", "birth_date", "age":
		order = "ji.birth_date IS NULL, ji.birth_date DESC, ji.name ASC"
	case "height":
		order = "ji.height_cm IS NULL, ji.height_cm ASC, ji.name ASC"
	case "bust":
		order = "ji.bust IS NULL, ji.bust DESC, ji.name ASC"
	case "hips", "hip":
		order = "ji.hips IS NULL, ji.hips DESC, ji.name ASC"
	case "waist":
		order = "ji.waist IS NULL, ji.waist ASC, ji.name ASC"
	case "measurements", "measure", "bwh":
		order = "ji.bust IS NULL, ji.bust DESC, ji.hips IS NULL, ji.hips DESC, ji.waist IS NULL, ji.waist ASC, ji.name ASC"
	case "cup":
		order = "ji.cup IS NULL, ji.cup DESC, ji.name ASC"
	case "work", "work_count", "count", "":
		// default order
	default:
		// ignore unknown values
	}
	if err := base.
		Select("ji.id, ji.name, ji.roman_name, ji.japanese_name, ji.chinese_name, ji.height_cm, ji.birth_date, ji.bust, ji.waist, ji.hips, ji.cup, COUNT(DISTINCT j.id) AS work_count, " + soloCode).
		Group("ji.id, ji.name, ji.roman_name, ji.japanese_name, ji.chinese_name, ji.height_cm, ji.birth_date, ji.bust, ji.waist, ji.hips, ji.cup").
		Order(order).
		Limit(limit).
		Offset(offset).
		Scan(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("list jav idols: %w", err)
	}

	return items, total, nil
}

// ListIdolCoverCodes returns a prioritized list of codes for an idol, preferring solo works first.
func ListIdolCoverCodes(ctx context.Context, idolID int64) ([]string, error) {
	var codes []string
	sub := common.DB.WithContext(ctx).
		Table("jav_idol_map").
		Select("jav_id, COUNT(*) as c").
		Group("jav_id")

	rows, err := common.DB.WithContext(ctx).
		Table("jav_idol_map jim").
		Select("j.code, CASE WHEN s.c = 1 THEN 1 ELSE 0 END as solo").
		Joins("JOIN jav j ON j.id = jim.jav_id").
		Joins("LEFT JOIN (?) s ON s.jav_id = jim.jav_id", sub).
		Where("jim.jav_idol_id = ?", idolID).
		Order("solo DESC, j.code").
		Rows()
	if err != nil {
		return nil, fmt.Errorf("list idol codes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var code string
		var solo int
		if err := rows.Scan(&code, &solo); err != nil {
			return nil, fmt.Errorf("scan idol codes: %w", err)
		}
		code = strings.TrimSpace(code)
		if code != "" {
			codes = append(codes, code)
		}
	}
	return codes, nil
}

// FindIdolSoloCode returns one solo work code for the idol, when available.
func FindIdolSoloCode(ctx context.Context, idolID int64) (string, error) {
	if idolID == 0 {
		return "", errors.New("idol id cannot be zero")
	}
	sub := common.DB.WithContext(ctx).
		Table("jav_idol_map").
		Select("jav_id, COUNT(*) as c").
		Group("jav_id")

	var codes []string
	if err := common.DB.WithContext(ctx).
		Table("jav_idol_map jim").
		Select("j.code").
		Joins("JOIN jav j ON j.id = jim.jav_id").
		Joins("LEFT JOIN (?) s ON s.jav_id = jim.jav_id", sub).
		Where("jim.jav_idol_id = ?", idolID).
		Where("s.c = 1").
		Order("RANDOM()").
		Limit(1).
		Pluck("j.code", &codes).Error; err != nil {
		return "", fmt.Errorf("find idol solo code: %w", err)
	}
	if len(codes) == 0 {
		return "", nil
	}
	return strings.TrimSpace(codes[0]), nil
}

// ListIdolsMissingProfile returns idols that have no profile fields populated.
func ListIdolsMissingProfile(ctx context.Context) ([]models.JavIdol, error) {
	var idols []models.JavIdol
	if err := common.DB.WithContext(ctx).
		Where(`
(
  japanese_name IS NULL OR japanese_name = '' OR
  roman_name IS NULL OR roman_name = '' OR
  chinese_name IS NULL OR chinese_name = '' OR
  height_cm IS NULL OR
  birth_date IS NULL OR
  bust IS NULL OR
  waist IS NULL OR
  hips IS NULL OR
  cup IS NULL
)`).
		Order("id").
		Find(&idols).Error; err != nil {
		return nil, fmt.Errorf("list idols missing profile: %w", err)
	}
	return idols, nil
}

// UpdateIdolProfile updates missing idol profile fields with fetched info.
func UpdateIdolProfile(ctx context.Context, idolID int64, info *jav.ActressInfo) error {
	if idolID == 0 {
		return errors.New("idol id cannot be zero")
	}
	if info == nil {
		return errors.New("actress info is nil")
	}
	updates := make(map[string]any)
	if name := strings.TrimSpace(info.JapaneseName); name != "" {
		updates["japanese_name"] = gorm.Expr("CASE WHEN japanese_name IS NULL OR japanese_name = '' THEN ? ELSE japanese_name END", name)
	}
	if name := strings.TrimSpace(info.RomanName); name != "" {
		updates["roman_name"] = gorm.Expr("CASE WHEN roman_name IS NULL OR roman_name = '' THEN ? ELSE roman_name END", name)
	}
	if name := strings.TrimSpace(info.ChineseName); name != "" {
		updates["chinese_name"] = gorm.Expr("CASE WHEN chinese_name IS NULL OR chinese_name = '' THEN ? ELSE chinese_name END", name)
	}
	if info.HeightCM > 0 {
		updates["height_cm"] = gorm.Expr("CASE WHEN height_cm IS NULL THEN ? ELSE height_cm END", info.HeightCM)
	}
	if info.BirthDate > 0 {
		updates["birth_date"] = gorm.Expr("CASE WHEN birth_date IS NULL THEN ? ELSE birth_date END", time.Unix(int64(info.BirthDate), 0).UTC())
	}
	if info.Bust > 0 {
		updates["bust"] = gorm.Expr("CASE WHEN bust IS NULL THEN ? ELSE bust END", info.Bust)
	}
	if info.Waist > 0 {
		updates["waist"] = gorm.Expr("CASE WHEN waist IS NULL THEN ? ELSE waist END", info.Waist)
	}
	if info.Hips > 0 {
		updates["hips"] = gorm.Expr("CASE WHEN hips IS NULL THEN ? ELSE hips END", info.Hips)
	}
	if info.Cup > 0 {
		updates["cup"] = gorm.Expr("CASE WHEN cup IS NULL THEN ? ELSE cup END", info.Cup)
	}
	if len(updates) == 0 {
		return nil
	}
	if err := common.DB.WithContext(ctx).
		Model(&models.JavIdol{}).
		Where("id = ?", idolID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("update idol profile: %w", err)
	}
	return nil
}

// ListVideosForJavScan loads fields used by the jav scanner.
func ListVideosForJavScan(ctx context.Context) ([]JavScanVideo, error) {
	var videos []JavScanVideo
	if err := common.DB.WithContext(ctx).
		Table("video").
		Joins("LEFT JOIN jav ON jav.id = video.jav_id").
		Where("COALESCE(hidden, 0) = 0").
		Select("video.id, video.filename, video.jav_id, video.updated_at, video.duration_sec, COALESCE(jav.provider, 0) AS jav_provider").
		Find(&videos).Error; err != nil {
		return nil, fmt.Errorf("list videos for jav scan: %w", err)
	}
	return videos, nil
}

// GetJavByCode fetches a jav record by code.
func GetJavByCode(ctx context.Context, code string) (*models.Jav, error) {
	var javRec models.Jav
	err := common.DB.WithContext(ctx).Where("code = ?", code).First(&javRec).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get jav by code: %w", err)
	}
	return &javRec, nil
}

// SetVideoJavID links a video to a jav record, guarding against stale updates when expectedUpdatedAt is provided.
func SetVideoJavID(ctx context.Context, videoID, javID int64, expectedUpdatedAt time.Time) error {
	return setVideoJavIDTx(common.DB.WithContext(ctx), videoID, javID, expectedUpdatedAt)
}

// SaveJavInfoAndLinkVideo upserts jav metadata and associates the video in one transaction.
func SaveJavInfoAndLinkVideo(ctx context.Context, info *jav.Info, videoID int64, expectedUpdatedAt time.Time) (*models.Jav, error) {
	if info == nil {
		return nil, errors.New("jav info is nil")
	}
	var javRec *models.Jav
	err := common.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		rec, err := saveJavInfoTx(tx, info)
		if err != nil {
			return err
		}
		if err := setVideoJavIDTx(tx, videoID, rec.ID, expectedUpdatedAt); err != nil {
			return err
		}
		javRec = rec
		return nil
	})
	if err != nil {
		return nil, err
	}
	return javRec, nil
}

// DeleteOrphanJavs removes JAV records that have no video referencing them.
func DeleteOrphanJavs(ctx context.Context) error {
	var orphanIDs []int64
	sub := common.DB.WithContext(ctx).Model(&models.Video{}).Select("DISTINCT jav_id").Where("jav_id IS NOT NULL")
	if err := common.DB.WithContext(ctx).Model(&models.Jav{}).
		Where("id NOT IN (?)", sub).
		Pluck("id", &orphanIDs).Error; err != nil {
		return fmt.Errorf("find orphan javs: %w", err)
	}
	if len(orphanIDs) == 0 {
		return nil
	}

	return common.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("jav_id IN ?", orphanIDs).Delete(&models.JavTagMap{}).Error; err != nil {
			return fmt.Errorf("delete orphan jav tag maps: %w", err)
		}
		if err := tx.Where("jav_id IN ?", orphanIDs).Delete(&models.JavIdolMap{}).Error; err != nil {
			return fmt.Errorf("delete orphan jav idol maps: %w", err)
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Where("id IN ?", orphanIDs).Delete(&models.Jav{}).Error; err != nil {
			return fmt.Errorf("delete orphan javs: %w", err)
		}
		return nil
	})
}

// ListJavCodes returns all jav codes.
func ListJavCodes(ctx context.Context) ([]string, error) {
	var codes []string
	if err := common.DB.WithContext(ctx).Model(&models.Jav{}).Pluck("code", &codes).Error; err != nil {
		return nil, fmt.Errorf("list jav codes: %w", err)
	}
	return codes, nil
}

func saveJavInfoTx(tx *gorm.DB, info *jav.Info, now ...time.Time) (*models.Jav, error) {
	if tx == nil {
		return nil, errors.New("tx is nil")
	}
	ts := time.Now()
	if len(now) > 0 {
		ts = now[0]
	}

	javRec, err := lockJavByCodeTx(tx, info.Code)
	if err != nil {
		return nil, err
	}
	if javRec == nil {
		javRec = &models.Jav{Code: info.Code}
	}
	javRec.Code = info.Code
	javRec.Title = info.Title
	javRec.ReleaseUnix = info.ReleaseUnix
	javRec.DurationMin = info.DurationMin
	javRec.Provider = int(jav.ParseProvider(int(info.Provider)))
	javRec.FetchedAt = ts
	if err := tx.Save(javRec).Error; err != nil {
		return nil, fmt.Errorf("save jav: %w", err)
	}

	tags, err := ensureJavTagsTx(tx, info.Tags, info.Provider)
	if err != nil {
		return nil, err
	}
	idols, err := ensureJavIdolsTx(tx, info.Actors)
	if err != nil {
		return nil, err
	}
	if err := tx.Model(javRec).Association("Tags").Replace(tags); err != nil {
		return nil, fmt.Errorf("replace jav tags: %w", err)
	}
	if err := tx.Model(javRec).Association("Idols").Replace(idols); err != nil {
		return nil, fmt.Errorf("replace jav idols: %w", err)
	}
	return javRec, nil
}

func lockJavByCodeTx(tx *gorm.DB, code string) (*models.Jav, error) {
	var javRec models.Jav
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("code = ?", code).First(&javRec).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get jav by code: %w", err)
	}
	return &javRec, nil
}

func ensureJavTagsTx(tx *gorm.DB, names []string, provider jav.Provider) ([]models.JavTag, error) {
	unique := normalizeNames(names)
	if len(unique) == 0 {
		return nil, nil
	}
	provider = jav.ParseProvider(int(provider))
	if provider == jav.ProviderUnknown {
		provider = jav.ProviderJavBus
	}
	var tags []models.JavTag
	for _, name := range unique {
		tag := models.JavTag{Name: name, Provider: int(provider)}
		if err := tx.Where("name = ? AND provider = ?", name, int(provider)).FirstOrCreate(&tag).Error; err != nil {
			return nil, fmt.Errorf("ensure jav tag %q: %w", name, err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func isUserJavTagProvider(provider int) bool {
	return jav.ParseProvider(provider) == jav.ProviderUser
}

func ensureJavIdolsTx(tx *gorm.DB, names []string) ([]models.JavIdol, error) {
	unique := normalizeNames(names)
	if len(unique) == 0 {
		return nil, nil
	}
	var idols []models.JavIdol
	for _, name := range unique {
		idol := models.JavIdol{Name: name}
		if err := tx.Where("name = ?", name).FirstOrCreate(&idol).Error; err != nil {
			return nil, fmt.Errorf("ensure jav idol %q: %w", name, err)
		}
		idols = append(idols, idol)
	}
	return idols, nil
}

func setVideoJavIDTx(tx *gorm.DB, videoID, javID int64, expectedUpdatedAt time.Time) error {
	if tx == nil {
		return errors.New("tx is nil")
	}
	q := tx.Model(&models.Video{}).Where("id = ?", videoID)
	if !expectedUpdatedAt.IsZero() {
		q = q.Where("updated_at = ?", expectedUpdatedAt)
	}
	res := q.Update("jav_id", javID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 && !expectedUpdatedAt.IsZero() {
		return fmt.Errorf("video %d stale or missing", videoID)
	}
	return nil
}
