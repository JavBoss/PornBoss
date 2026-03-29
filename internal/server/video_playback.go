package server

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"pornboss/internal/common/logging"
	dbpkg "pornboss/internal/db"
	"pornboss/internal/models"
	"pornboss/internal/util"
)

const (
	hlsManifestMime = "application/vnd.apple.mpegurl"
	hlsSegmentMime  = "video/MP2T"
	hlsWaitTimeout  = 20 * time.Second
	hlsPollInterval = 200 * time.Millisecond
	hlsSegmentSecs  = 4
	hlsBufferCount  = 8
)

var hlsSegmentNamePattern = regexp.MustCompile(`^\d+\.ts$`)

type playbackResponse struct {
	Mode string `json:"mode"`
	Src  string `json:"src"`
	Type string `json:"type"`
}

type hlsManager struct {
	mu       sync.Mutex
	sessions map[string]*hlsSession
}

type hlsSession struct {
	inputPath    string
	dir          string
	manifestPath string
	durationSec  int64
	mu           sync.Mutex
	process      *hlsProcess
}

type hlsProcess struct {
	startSegment int
	endSegment   int
	done         chan struct{}
	cancel       context.CancelFunc
	err          error
}

var videoHLSManager = &hlsManager{
	sessions: make(map[string]*hlsSession),
}

func getVideoPlayback(c *gin.Context) {
	video, fullPath, ok := loadVideoByID(c)
	if !ok {
		return
	}
	if err := ensureVideoFileExists(c, fullPath); err != nil {
		return
	}

	meta, err := util.ProbeVideo(fullPath)
	if err != nil {
		logging.Error("probe playback metadata error: %v", err)
	}

	mode, mimeType := util.DetermineBrowserPlayback(fullPath, meta)
	if mode == util.BrowserPlaybackModeHLS {
		if _, err := util.ResolveFFmpegPath(); err != nil {
			logging.Error("resolve ffmpeg for hls playback error: %v", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ffmpeg unavailable for browser playback"})
			return
		}
	}

	src := fmt.Sprintf("/videos/%d/stream", video.ID)
	if mode == util.BrowserPlaybackModeHLS {
		src = fmt.Sprintf("/videos/%d/stream.m3u8", video.ID)
	}

	c.JSON(http.StatusOK, playbackResponse{
		Mode: mode,
		Src:  src,
		Type: mimeType,
	})
}

func streamVideoManifest(c *gin.Context) {
	video, fullPath, ok := loadVideoByID(c)
	if !ok {
		return
	}
	if err := ensureVideoFileExists(c, fullPath); err != nil {
		return
	}

	if _, err := videoHLSManager.sessionForFile(fullPath, video.DurationSec); err != nil {
		logging.Error("prepare hls session error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "prepare hls stream failed"})
		return
	}

	manifest, err := buildStaticHLSManifest(video.DurationSec)
	if err != nil {
		logging.Error("build hls manifest error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "build hls manifest failed"})
		return
	}

	c.Header("Content-Type", hlsManifestMime)
	c.Header("Cache-Control", "no-cache")
	c.String(http.StatusOK, manifest)
}

func streamVideoSegment(c *gin.Context) {
	video, fullPath, ok := loadVideoByID(c)
	if !ok {
		return
	}
	if err := ensureVideoFileExists(c, fullPath); err != nil {
		return
	}

	segment := strings.TrimSpace(c.Param("segment"))
	if !hlsSegmentNamePattern.MatchString(segment) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid segment"})
		return
	}

	session, err := videoHLSManager.sessionForFile(fullPath, video.DurationSec)
	if err != nil {
		logging.Error("prepare hls session error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "prepare hls stream failed"})
		return
	}

	segmentIndex, _ := strconv.Atoi(strings.TrimSuffix(segment, ".ts"))
	segmentPath := filepath.Join(session.dir, segment)
	if err := session.ensureSegment(c.Request.Context(), segmentIndex, segmentPath, hlsWaitTimeout); err != nil {
		writeHLSError(c, err)
		return
	}

	c.Header("Content-Type", hlsSegmentMime)
	c.Header("Cache-Control", "public, max-age=300")
	c.File(segmentPath)
}

func loadVideoByID(c *gin.Context) (*models.Video, string, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return nil, "", false
	}

	video, err := dbpkg.GetVideo(c.Request.Context(), id)
	if err != nil {
		logging.Error("get video error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return nil, "", false
	}
	if video == nil {
		c.Status(http.StatusNotFound)
		return nil, "", false
	}

	fullPath, _, err := resolveVideoPath(video.Path, video.DirectoryRef.Path)
	if err != nil {
		logging.Error("resolve video path by id error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return nil, "", false
	}

	return video, fullPath, true
}

func (m *hlsManager) sessionForFile(fullPath string, durationSec int64) (*hlsSession, error) {
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	key := buildHLSCacheKey(fullPath, info)
	baseDir := filepath.Join(os.TempDir(), "pornboss-hls", key)

	m.mu.Lock()
	session := m.sessions[key]
	if session == nil {
		session = &hlsSession{
			inputPath:    fullPath,
			dir:          baseDir,
			manifestPath: filepath.Join(baseDir, "manifest.m3u8"),
			durationSec:  durationSec,
		}
		m.sessions[key] = session
	} else if durationSec > 0 && session.durationSec != durationSec {
		session.durationSec = durationSec
	}
	m.mu.Unlock()

	if err := os.MkdirAll(session.dir, 0o755); err != nil {
		return nil, err
	}
	return session, nil
}

func buildHLSCacheKey(fullPath string, info os.FileInfo) string {
	sum := sha1.Sum([]byte(fullPath + "|" + strconv.FormatInt(info.Size(), 10) + "|" + strconv.FormatInt(info.ModTime().UTC().UnixNano(), 10)))
	return hex.EncodeToString(sum[:])
}

func (s *hlsSession) ensureSegment(ctx context.Context, segmentIndex int, path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if ctx != nil {
			if err := ctx.Err(); err != nil {
				return err
			}
		}
		if fileExists(path) {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for hls output: %s", filepath.Base(path))
		}

		proc, err := s.ensureProcessForSegment(ctx, segmentIndex)
		if err != nil {
			return err
		}

		timer := time.NewTimer(hlsPollInterval)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()
		case <-proc.done:
			if !timer.Stop() {
				<-timer.C
			}
			if fileExists(path) {
				return nil
			}
			if proc.err != nil {
				return proc.err
			}
		case <-timer.C:
		}
	}
}

func (s *hlsSession) ensureProcessForSegment(ctx context.Context, segmentIndex int) (*hlsProcess, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}

	if s.process != nil {
		if s.process.coversSegment(segmentIndex) {
			return s.process, nil
		}
		s.process.cancel()
		s.process = nil
	}

	proc, err := s.startProcessLocked(segmentIndex)
	if err != nil {
		return nil, err
	}
	s.process = proc
	return proc, nil
}

func (s *hlsSession) startProcessLocked(segmentIndex int) (*hlsProcess, error) {
	ctx, cancel := context.WithCancel(context.Background())
	bufferSeconds := (hlsBufferCount + 1) * hlsSegmentSecs
	if s.durationSec > 0 {
		startOffset := segmentIndex * hlsSegmentSecs
		remaining := int(s.durationSec) - startOffset
		if remaining > 0 && remaining < bufferSeconds {
			bufferSeconds = remaining
		}
	}

	proc := &hlsProcess{
		startSegment: segmentIndex,
		endSegment:   segmentIndex + hlsSegmentWindowSize(bufferSeconds) - 1,
		done:         make(chan struct{}),
		cancel:       cancel,
	}

	ffmpegPath, err := util.ResolveFFmpegPath()
	if err != nil {
		cancel()
		return nil, err
	}

	startOffset := segmentIndex * hlsSegmentSecs

	args := []string{
		"-hide_banner",
		"-loglevel", "error",
		"-nostdin",
		"-y",
		"-ss", strconv.Itoa(startOffset),
		"-i", s.inputPath,
		"-map", "0:v:0",
		"-map", "0:a?",
		"-t", strconv.Itoa(bufferSeconds),
		"-c:v", "libx264",
		"-preset", "veryfast",
		"-crf", "23",
		"-force_key_frames", fmt.Sprintf("expr:gte(t,n_forced*%d)", hlsSegmentSecs),
		"-c:a", "aac",
		"-ac", "2",
		"-sn",
		"-f", "hls",
		"-start_number", strconv.Itoa(segmentIndex),
		"-hls_time", strconv.Itoa(hlsSegmentSecs),
		"-hls_list_size", "0",
		"-hls_playlist_type", "vod",
		"-hls_base_url", "stream.m3u8/",
		"-hls_flags", "split_by_time+independent_segments+temp_file",
		"-hls_segment_filename", filepath.Join(s.dir, "%06d.ts"),
		s.manifestPath,
	}

	cmd := exec.CommandContext(ctx, ffmpegPath, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	go func() {
		runErr := cmd.Run()
		if runErr != nil && !errors.Is(runErr, context.Canceled) {
			msg := strings.TrimSpace(stderr.String())
			if msg != "" {
				proc.err = fmt.Errorf("ffmpeg hls failed: %w: %s", runErr, msg)
			} else {
				proc.err = fmt.Errorf("ffmpeg hls failed: %w", runErr)
			}
		}

		s.mu.Lock()
		if s.process == proc {
			s.process = nil
		}
		s.mu.Unlock()
		close(proc.done)
	}()

	return proc, nil
}

func (p *hlsProcess) coversSegment(segmentIndex int) bool {
	if p == nil {
		return false
	}
	return segmentIndex >= p.startSegment && segmentIndex <= p.endSegment
}

func hlsSegmentWindowSize(bufferSeconds int) int {
	if bufferSeconds <= 0 {
		return 1
	}
	windowSize := int(math.Ceil(float64(bufferSeconds) / float64(hlsSegmentSecs)))
	if windowSize < 1 {
		return 1
	}
	return windowSize
}

func buildStaticHLSManifest(durationSec int64) (string, error) {
	if durationSec <= 0 {
		return "", errors.New("invalid video duration")
	}

	segmentCount := int(math.Ceil(float64(durationSec) / float64(hlsSegmentSecs)))
	if segmentCount < 1 {
		segmentCount = 1
	}

	var buf strings.Builder
	buf.WriteString("#EXTM3U\n")
	buf.WriteString("#EXT-X-VERSION:3\n")
	buf.WriteString("#EXT-X-PLAYLIST-TYPE:VOD\n")
	buf.WriteString(fmt.Sprintf("#EXT-X-TARGETDURATION:%d\n", hlsSegmentSecs))
	buf.WriteString("#EXT-X-MEDIA-SEQUENCE:0\n")

	remaining := float64(durationSec)
	for i := 0; i < segmentCount; i++ {
		segmentDuration := float64(hlsSegmentSecs)
		if remaining > 0 && remaining < segmentDuration {
			segmentDuration = remaining
		}
		if segmentDuration <= 0 {
			segmentDuration = float64(hlsSegmentSecs)
		}

		buf.WriteString(fmt.Sprintf("#EXTINF:%.3f,\n", segmentDuration))
		buf.WriteString(fmt.Sprintf("stream.m3u8/%06d.ts\n", i))
		remaining -= segmentDuration
	}

	buf.WriteString("#EXT-X-ENDLIST\n")
	return buf.String(), nil
}

func writeHLSError(c *gin.Context, err error) {
	if errors.Is(err, context.Canceled) {
		return
	}
	logging.Error("serve hls error: %v", err)
	if errors.Is(err, os.ErrNotExist) {
		c.Status(http.StatusNotFound)
		return
	}
	if strings.Contains(err.Error(), "timeout waiting for hls output") {
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "hls stream timeout"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "hls stream failed"})
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
