package server

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"pornboss/internal/common"
	"pornboss/internal/common/logging"
	dbpkg "pornboss/internal/db"
	"pornboss/internal/manager"
	"pornboss/internal/models"
	"pornboss/internal/mpv"
	"pornboss/internal/util"
)

func listVideos(c *gin.Context) {
	limit := queryInt(c, "limit", 100)
	offset := queryInt(c, "offset", 0)
	tagFilter := parseTagQuery(c.Query("tags"))
	search := strings.TrimSpace(c.Query("search"))
	sort := strings.TrimSpace(c.Query("sort"))
	seedParam := strings.TrimSpace(c.Query("seed"))
	var seed *int64
	if seedParam != "" {
		parsed, err := strconv.ParseInt(seedParam, 10, 64)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid seed"})
			return
		}
		seed = &parsed
	}

	videos, err := dbpkg.ListVideos(c.Request.Context(), limit, offset, tagFilter, search, sort, seed)
	if err != nil {
		logging.Error("list videos error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	total, err := dbpkg.CountVideos(c.Request.Context(), tagFilter, search)
	if err != nil {
		logging.Error("count videos error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": videos,
		"total": total,
	})
}

func getVideo(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	video, err := dbpkg.GetVideo(c.Request.Context(), id)
	if err != nil {
		logging.Error("get video error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if video == nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, video)
}

func incrementVideoPlayCount(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := dbpkg.IncrementVideoPlayCount(c.Request.Context(), id); err != nil {
		logging.Error("increment play count error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type playbackSource struct {
	Kind     string `json:"kind"`
	Src      string `json:"src"`
	MimeType string `json:"mime_type"`
	Label    string `json:"label"`
}

type playbackInfo struct {
	VideoID       int64            `json:"video_id"`
	PreferredKind string           `json:"preferred_kind"`
	Sources       []playbackSource `json:"sources"`
}

func getVideoStreams(c *gin.Context) {
	video, fullPath, err := resolveVideoStreamTarget(c)
	if err != nil {
		respondPlaybackError(c, err)
		return
	}

	probe, err := util.ProbePlaybackSupport(fullPath)
	if err != nil {
		logging.Error("probe playback support error: %v", err)
		respondPlaybackError(c, err)
		return
	}

	info := playbackInfo{
		VideoID:       video.ID,
		PreferredKind: "hls",
		Sources:       []playbackSource{},
	}
	if probe.SupportsDirect {
		info.PreferredKind = "direct"
		info.Sources = append(info.Sources, playbackSource{
			Kind:     "direct",
			Src:      buildDirectStreamURL(video),
			MimeType: directMimeType(probe.Container),
			Label:    "Direct",
		})
	}
	info.Sources = append(info.Sources, playbackSource{
		Kind:     "hls",
		Src:      "/videos/" + strconv.FormatInt(video.ID, 10) + "/stream.m3u8",
		MimeType: manager.MimeHLS,
		Label:    "HLS",
	})

	c.JSON(http.StatusOK, info)
}

func streamVideo(c *gin.Context) {
	fullPath, err := resolveStreamPathFromQuery(c)
	if err != nil {
		_, fullPath, err = resolveVideoStreamTarget(c)
		if err != nil {
			respondPlaybackError(c, err)
			return
		}
	}
	serveVideoFile(c, fullPath)
}

func streamHLSManifest(c *gin.Context) {
	video, fullPath, err := resolveVideoStreamTarget(c)
	if err != nil {
		respondPlaybackError(c, err)
		return
	}
	if common.StreamManager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "stream manager unavailable"})
		return
	}

	resolution := strings.TrimSpace(c.Query("resolution"))
	c.Header("Cache-Control", "no-cache")
	common.StreamManager.ServeManifest(c.Writer, c.Request, fullPath, float64(video.DurationSec), resolution)
}

func streamHLSSegment(c *gin.Context) {
	video, fullPath, err := resolveVideoStreamTarget(c)
	if err != nil {
		respondPlaybackError(c, err)
		return
	}
	if common.StreamManager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "stream manager unavailable"})
		return
	}

	segment := strings.TrimSpace(c.Param("segment"))
	resolution := strings.TrimSpace(c.Query("resolution"))
	c.Header("Cache-Control", "no-cache")
	common.StreamManager.ServeSegment(c.Writer, c.Request, manager.StreamOptions{
		StreamType: manager.StreamTypeHLS,
		SourcePath: fullPath,
		Duration:   float64(video.DurationSec),
		Resolution: resolution,
		Key:        strconv.FormatInt(video.ID, 10),
		Segment:    segment,
	})
}

func resolveStreamPathFromQuery(c *gin.Context) (string, error) {
	rawPath := strings.TrimSpace(c.Query("path"))
	rawDirPath := strings.TrimSpace(c.Query("dir_path"))
	fullPath, _, err := resolveVideoPath(rawPath, rawDirPath)
	return fullPath, err
}

func resolveVideoStreamTarget(c *gin.Context) (*models.Video, string, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		return nil, "", errors.New("invalid id")
	}

	video, err := dbpkg.GetVideo(c.Request.Context(), id)
	if err != nil {
		return nil, "", err
	}
	if video == nil {
		return nil, "", os.ErrNotExist
	}

	fullPath, _, err := resolveVideoPath(video.Path, video.DirectoryRef.Path)
	if err != nil {
		return nil, "", err
	}
	if _, err := os.Stat(fullPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, "", err
		}
		return nil, "", err
	}

	return video, fullPath, nil
}

func respondPlaybackError(c *gin.Context, err error) {
	switch {
	case err == nil:
		return
	case errors.Is(err, os.ErrNotExist):
		c.Status(http.StatusNotFound)
	case errors.Is(err, context.Canceled):
		c.Status(499)
	case strings.Contains(err.Error(), "ffmpeg not found"), strings.Contains(err.Error(), "ffprobe not found"):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
	case strings.Contains(err.Error(), "invalid segment"), strings.Contains(err.Error(), "invalid id"), strings.Contains(err.Error(), "invalid path"):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}

func directMimeType(container string) string {
	switch strings.ToLower(strings.TrimSpace(container)) {
	case "webm":
		return "video/webm"
	default:
		return "video/mp4"
	}
}

func buildDirectStreamURL(video *models.Video) string {
	if video == nil {
		return ""
	}
	base := "/videos/" + strconv.FormatInt(video.ID, 10) + "/stream"
	values := url.Values{}
	if path := strings.TrimSpace(video.Path); path != "" {
		values.Set("path", path)
	}
	if dirPath := strings.TrimSpace(video.DirectoryRef.Path); dirPath != "" {
		values.Set("dir_path", dirPath)
	}
	if len(values) == 0 {
		return base
	}
	return base + "?" + values.Encode()
}

func serveVideoFile(c *gin.Context, fullPath string) {
	if _, err := os.Stat(fullPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.Status(http.StatusNotFound)
			return
		}
		logging.Error("stat stream file error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.File(fullPath)
}

func openVideoFile(c *gin.Context) {
	fullPath, dirPath, err := resolveVideoPathFromBody(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := ensureVideoFileExists(c, fullPath); err != nil {
		return
	}
	if err := util.OpenFile(fullPath); err != nil {
		logging.Error("open video file error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "open file failed"})
		return
	}
	incrementPlayCountByPath(c.Request.Context(), dirPath, fullPath)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func playVideoFile(c *gin.Context) {
	fullPath, dirPath, err := resolveVideoPathFromBody(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := ensureVideoFileExists(c, fullPath); err != nil {
		return
	}
	if err := mpv.PlayVideo(fullPath); err != nil {
		logging.Error("play video file error: %v", err)
		if strings.Contains(err.Error(), "mpv not found") {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "play file failed"})
		return
	}
	incrementPlayCountByPath(c.Request.Context(), dirPath, fullPath)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func revealVideoLocation(c *gin.Context) {
	fullPath, _, err := resolveVideoPathFromBody(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := ensureVideoFileExists(c, fullPath); err != nil {
		return
	}
	if err := util.RevealFile(fullPath); err != nil {
		logging.Error("reveal video file error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "reveal file failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type videoPathRequest struct {
	Path    string `json:"path"`
	DirPath string `json:"dir_path"`
}

func resolveVideoPathFromBody(c *gin.Context) (string, string, error) {
	var req videoPathRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return "", "", errors.New("invalid payload")
	}
	fullPath, dirPath, err := resolveVideoPath(req.Path, req.DirPath)
	return fullPath, dirPath, err
}

func resolveVideoPath(rawPath, rawDirPath string) (string, string, error) {
	if strings.TrimSpace(rawPath) == "" || strings.TrimSpace(rawDirPath) == "" {
		return "", "", errors.New("path and dir_path are required")
	}

	dirPath := filepath.Clean(rawDirPath)
	if dirPath == "." || !filepath.IsAbs(dirPath) {
		return "", "", errors.New("invalid dir_path")
	}

	cleanPath := filepath.Clean(filepath.FromSlash(rawPath))
	if cleanPath == "." {
		return "", "", errors.New("invalid path")
	}

	fullPath := cleanPath
	if !filepath.IsAbs(cleanPath) {
		fullPath = filepath.Join(dirPath, cleanPath)
	}

	relCheck, err := filepath.Rel(dirPath, fullPath)
	if err != nil || relCheck == ".." || strings.HasPrefix(relCheck, ".."+string(os.PathSeparator)) {
		return "", "", errors.New("invalid path")
	}
	return fullPath, dirPath, nil
}

func ensureVideoFileExists(c *gin.Context, fullPath string) error {
	if _, err := os.Stat(fullPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.Status(http.StatusNotFound)
			return err
		}
		logging.Error("stat stream file error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return err
	}
	return nil
}

func incrementPlayCountByPath(ctx context.Context, dirPath, fullPath string) {
	if strings.TrimSpace(dirPath) == "" || strings.TrimSpace(fullPath) == "" {
		return
	}
	relPath, err := filepath.Rel(dirPath, fullPath)
	if err != nil {
		logging.Error("resolve relative path for play count: %v", err)
		return
	}
	relPath = filepath.ToSlash(filepath.Clean(relPath))
	if relPath == "." || strings.HasPrefix(relPath, "..") {
		return
	}
	if err := dbpkg.IncrementVideoPlayCountByPath(ctx, dirPath, relPath); err != nil {
		logging.Error("increment play count by path error: %v", err)
	}
}

func getThumbnail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	video, err := dbpkg.GetVideo(c.Request.Context(), id)
	if err != nil {
		logging.Error("get screenshot error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if video == nil {
		c.Status(http.StatusNotFound)
		return
	}

	second, ok := manager.PickScreenshotSecond(video.DurationSec)
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}

	if common.AppConfig == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	dataDir := filepath.Dir(common.AppConfig.DatabasePath)
	screenshotPath := manager.ScreenshotPath(dataDir, video.ID, second)
	if screenshotPath == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if _, err := os.Stat(screenshotPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			common.ScreenshotManager.EnqueueForVideo(video)
			c.Status(http.StatusNotFound)
			return
		}
		logging.Error("stat screenshot error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.File(screenshotPath)
}
