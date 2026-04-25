package manager

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"pornboss/internal/common/logging"
	"pornboss/internal/models"
	"pornboss/internal/util"
)

// Task represents a request to capture a screenshot for a specific video.
type Task struct {
	VideoID    int64
	Second     int
	ModifiedAt time.Time
	Size       int64
}

// VideoFetcher loads a video record by ID.
type VideoFetcher func(ctx context.Context, id int64) (*models.Video, error)

const maxScreenshotWorkers = 8

// ScreenshotManager coordinates asynchronous screenshot generation using the worker.
type ScreenshotManager struct {
	tasks      chan Task
	workers    int
	dataDir    string
	fetchVideo VideoFetcher
}

// NewScreenshotManager creates a manager when dataDir and fetchVideo are provided.
// Returns nil when either is missing, effectively disabling screenshot generation.
func NewScreenshotManager(dataDir string, fetchVideo VideoFetcher) *ScreenshotManager {
	dataDir = strings.TrimSpace(dataDir)
	if dataDir == "" || fetchVideo == nil {
		return nil
	}
	workers := runtime.GOMAXPROCS(0)
	if workers <= 0 {
		workers = 1
	}
	if workers > maxScreenshotWorkers {
		workers = maxScreenshotWorkers
	}
	logging.Info("screenshot manager initialized with %d workers", workers)
	return &ScreenshotManager{
		tasks:      make(chan Task, 5000),
		workers:    workers,
		dataDir:    dataDir,
		fetchVideo: fetchVideo,
	}
}

// Start launches the background worker. Safe to call with nil manager.
func (m *ScreenshotManager) Start(ctx context.Context) {
	if m == nil {
		return
	}
	if m.workers <= 0 {
		m.workers = runtime.GOMAXPROCS(0)
		if m.workers <= 0 {
			m.workers = 1
		}
		if m.workers > maxScreenshotWorkers {
			m.workers = maxScreenshotWorkers
		}
	}
	for i := 0; i < m.workers; i++ {
		go m.startWorker(ctx)
	}
}

// Enqueue schedules a single screenshot task. Invalid or empty tasks are ignored.
func (m *ScreenshotManager) Enqueue(task Task) {
	if m == nil {
		return
	}
	if task.VideoID <= 0 || task.Second <= 0 || task.ModifiedAt.IsZero() {
		return
	}
	// Block until the worker takes it to ensure screenshots are always generated.
	m.tasks <- task
}

// EnqueueForVideo schedules a screenshot task using the standard second selection logic.
func (m *ScreenshotManager) EnqueueForVideo(video *models.Video) {
	if m == nil {
		return
	}
	task, ok := TaskForVideo(video)
	if !ok {
		return
	}
	m.Enqueue(task)
}

// ScreenshotPath builds the on-disk screenshot path for a video ID and second.
func (m *ScreenshotManager) ScreenshotPath(videoID int64, second int) string {
	if m == nil {
		return ""
	}
	return ScreenshotPath(m.dataDir, videoID, second)
}

// ScreenshotPath builds the on-disk screenshot path for a video ID and second.
func ScreenshotPath(dataDir string, videoID int64, second int) string {
	dataDir = strings.TrimSpace(dataDir)
	if dataDir == "" || videoID <= 0 || second <= 0 {
		return ""
	}
	fileName := fmt.Sprintf("%d.jpg", second)
	return filepath.Join(dataDir, "video", strconv.FormatInt(videoID, 10), "screenshot", fileName)
}

var screenshotSeconds = []int{128, 63, 32, 16, 8, 4, 2, 1}

// PickScreenshotSecond picks the closest configured second that does not exceed durationSec.
func PickScreenshotSecond(durationSec int64) (int, bool) {
	if durationSec <= 0 {
		return 0, false
	}
	for _, candidate := range screenshotSeconds {
		if durationSec >= int64(candidate) {
			return candidate, true
		}
	}
	return 0, false
}

// TaskForVideo builds a screenshot task for the given video using standard selection logic.
func TaskForVideo(video *models.Video) (Task, bool) {
	if video == nil {
		return Task{}, false
	}
	if video.ID <= 0 || video.ModifiedAt.IsZero() {
		return Task{}, false
	}
	second, ok := PickScreenshotSecond(video.DurationSec)
	if !ok {
		return Task{}, false
	}
	return Task{
		VideoID:    video.ID,
		Second:     second,
		ModifiedAt: video.ModifiedAt,
		Size:       video.Size,
	}, true
}

// startWorker launches a background loop that consumes screenshot generation
// tasks. The worker stops when the context is done or the channel is closed.
func (m *ScreenshotManager) startWorker(ctx context.Context) {
	if m == nil || m.dataDir == "" || m.fetchVideo == nil {
		logging.Info("screenshot worker disabled: data directory or fetcher missing")
		return
	}

	for {
		select {
		case <-ctx.Done():
			logging.Info("screenshot worker exiting: context cancelled")
			return
		case task, ok := <-m.tasks:
			if !ok {
				logging.Info("screenshot worker exiting: task channel closed")
				return
			}
			if err := m.processTask(ctx, task); err != nil {
				logging.Error("screenshot task failed (video_id=%d, second=%d): %v", task.VideoID, task.Second, err)
			}
		}
	}
}

func (m *ScreenshotManager) processTask(parent context.Context, task Task) error {
	if task.VideoID <= 0 || task.Second <= 0 || task.ModifiedAt.IsZero() {
		return errors.New("invalid screenshot task: missing video id, second, or modified_at")
	}
	screenshotPath := m.ScreenshotPath(task.VideoID, task.Second)
	if screenshotPath == "" {
		return errors.New("invalid screenshot task: missing screenshot path")
	}
	if _, err := os.Stat(screenshotPath); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat screenshot: %w", err)
	}

	video, err := m.fetchVideo(parent, task.VideoID)
	if err != nil {
		return err
	}
	if video == nil {
		return nil
	}
	if !sameVideoMeta(video.ModifiedAt, video.Size, task) {
		return nil
	}

	videoPath, err := resolveVideoPath(video)
	if err != nil {
		return err
	}

	info, err := os.Stat(videoPath)
	if err != nil {
		return err
	}
	if !sameVideoMeta(info.ModTime(), info.Size(), task) {
		return nil
	}

	// Bound mpv execution time to avoid stuck processes.
	ctx, cancel := context.WithTimeout(parent, 2*time.Minute)
	defer cancel()

	return m.capture(ctx, videoPath, task.Second, screenshotPath)
}

func (m *ScreenshotManager) capture(ctx context.Context, videoPath string, second int, outputPath string) error {
	if videoPath == "" {
		return errors.New("video path is required")
	}
	if second <= 0 {
		return errors.New("second is required")
	}
	if outputPath == "" {
		return errors.New("output path is required")
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("ensure screenshot dir: %w", err)
	}

	mpvPath, pathErr := util.ResolveMPVPath()
	if pathErr != nil {
		return fmt.Errorf("resolve mpv path: %w", pathErr)
	}
	tempDir, err := os.MkdirTemp(filepath.Dir(outputPath), ".mpv-shot-*")
	if err != nil {
		return fmt.Errorf("create mpv temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()
	shotPath := filepath.Join(tempDir, "00000001.jpg")

	args := []string{
		"--no-config",
		"--really-quiet",
		"--msg-level=all=error",
		"--ao=null",
		"--hr-seek=yes",
		"--start=" + strconv.Itoa(second),
		"--frames=1",
		"--vo=image",
		"--vo-image-format=jpg",
		"--vo-image-outdir=" + tempDir,
		videoPath,
	}

	cmd := exec.CommandContext(ctx, mpvPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		_ = os.Remove(shotPath)
		if errors.Is(err, exec.ErrNotFound) {
			return fmt.Errorf("mpv not found: %w", err)
		}
		lastOut := strings.TrimSpace(string(out))
		if lastOut != "" {
			return fmt.Errorf("mpv screenshot failed: %w: %s", err, lastOut)
		}
		return fmt.Errorf("mpv screenshot failed: %w", err)
	}

	info, err := os.Stat(shotPath)
	if err != nil {
		return errors.New("mpv produced no screenshot file")
	}
	if info.Size() == 0 {
		_ = os.Remove(shotPath)
		return errors.New("mpv produced empty screenshot file")
	}

	if err := os.Rename(shotPath, outputPath); err != nil {
		return fmt.Errorf("rename screenshot: %w", err)
	}
	return nil
}

func resolveVideoPath(video *models.Video) (string, error) {
	if video == nil {
		return "", errors.New("video is nil")
	}
	dirPath := strings.TrimSpace(video.DirectoryRef.Path)
	if dirPath == "" {
		return "", errors.New("video directory missing")
	}
	relPath := strings.TrimSpace(video.Path)
	if relPath == "" {
		return "", errors.New("video path missing")
	}
	return filepath.Join(dirPath, filepath.FromSlash(relPath)), nil
}

func sameVideoMeta(modifiedAt time.Time, size int64, task Task) bool {
	if task.ModifiedAt.IsZero() {
		return false
	}
	if size != task.Size {
		return false
	}
	return modifiedAt.Equal(task.ModifiedAt)
}
