package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/h2non/filetype"
)

type VideoMetadata struct {
	Codec           string
	AudioCodec      string
	Container       string
	Width           int
	Height          int
	FPS             float64
	SampleRate      int
	Channels        int
	DurationSeconds float64
	FormatBitRate   int64
	VideoBitRate    int64
	AudioBitRate    int64
}

func (m *VideoMetadata) Fingerprint(size int64) string {
	fps := m.FPS
	if fps > 0 {
		fps = math.Round(fps*1000) / 1000
	}
	dur := math.Round(m.DurationSeconds)
	return fmt.Sprintf("%dx%d|%s|%.3f|%d|%d|%.0f|%d",
		m.Width,
		m.Height,
		strings.TrimSpace(m.Codec),
		fps,
		m.SampleRate,
		m.Channels,
		dur,
		size)
}

// FingerprintV2 returns a metadata-only fingerprint with higher granularity.
// Format: widthxheight|bitrate|video_bitrate|audio_bitrate|duration_ms|size
func (m *VideoMetadata) FingerprintV2(size int64) string {
	durationMs := int64(math.Round(m.DurationSeconds * 1000))
	return fmt.Sprintf("%dx%d|%d|%d|%d|%d|%d",
		m.Width,
		m.Height,
		m.FormatBitRate,
		m.VideoBitRate,
		m.AudioBitRate,
		durationMs,
		size)
}

// isVideo uses github.com/h2non/filetype to detect if the file is a video by
// inspecting the initial bytes and matching known MIME signatures.
func IsVideo(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".ts" || ext == ".mts" || ext == ".m2ts" {
		f, err := os.Open(path)
		if err != nil {
			return false
		}
		defer f.Close()
		// MPEG-TS packets start with 0x47 sync byte every 188 bytes.
		header := make([]byte, 564) // 3 * 188
		n, _ := f.Read(header)
		if n < 188 {
			return false
		}
		if header[0] != 0x47 {
			return false
		}
		if n > 188 && header[188] != 0x47 {
			return false
		}
		if n > 376 && header[376] != 0x47 {
			return false
		}
		return true
	}

	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	// filetype recommends at least 261 bytes
	header := make([]byte, 261)
	n, err := f.Read(header)
	if n == 0 && err != nil {
		return false
	}
	buf := header[:n]
	kind, err := filetype.Match(buf)
	if err != nil {
		return false
	}
	if kind == filetype.Unknown {
		return false
	}
	// Accept any MIME with top-level type "video"
	return strings.HasPrefix(kind.MIME.Value, "video/") || kind.MIME.Type == "video"
}

var (
	ffprobeOnce sync.Once
	ffprobePath string
	ffprobeErr  error
)

// ResolveFFprobePath resolves the ffprobe binary location.
func ResolveFFprobePath() (string, error) {
	ffprobeOnce.Do(func() {
		ffprobePath, ffprobeErr = findFFprobePath()
	})
	return ffprobePath, ffprobeErr
}

func findFFprobePath() (string, error) {
	var candidates []string
	if env := strings.TrimSpace(os.Getenv("FFPROBE_PATH")); env != "" {
		candidates = append(candidates, env)
	}

	binName := "ffprobe"
	if runtime.GOOS == "windows" {
		binName = "ffprobe.exe"
	}

	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, "internal", "bin", binName))
	}
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		candidates = append(candidates, filepath.Join(execDir, "internal", "bin", binName))
	}
	candidates = append(candidates, binName)

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if resolved, err := exec.LookPath(candidate); err == nil {
			return resolved, nil
		}
	}
	return "", errors.New("ffprobe not found; set FFPROBE_PATH or place binary at internal/bin/ffprobe")
}

// ProbeVideo extracts codec/resolution/fps/duration using ffprobe.
func ProbeVideo(path string) (*VideoMetadata, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("empty path")
	}
	ffprobe, err := ResolveFFprobePath()
	if err != nil {
		return nil, err
	}
	// -v quiet -print_format json -show_streams -select_streams v:0
	cmd := exec.Command(ffprobe,
		"-v", "error",
		"-print_format", "json",
		"-show_entries", "stream=index,codec_type,codec_name,width,height,avg_frame_rate,r_frame_rate,sample_rate,channels,bit_rate:format=format_name,duration,size,bit_rate",
		path,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg != "" {
			return nil, fmt.Errorf("ffprobe: %w: %s", err, errMsg)
		}
		return nil, fmt.Errorf("ffprobe: %w", err)
	}
	meta, err := parseFFprobeOutput(out)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

type ffprobeStream struct {
	CodecName    string `json:"codec_name"`
	CodecType    string `json:"codec_type"`
	PixFmt       string `json:"pix_fmt"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	AvgFrameRate string `json:"avg_frame_rate"`
	RFrameRate   string `json:"r_frame_rate"`
	Duration     string `json:"duration"`
	DurationTS   int64  `json:"duration_ts"`
	SampleRate   string `json:"sample_rate"`
	Channels     int    `json:"channels"`
	BitRate      string `json:"bit_rate"`
}
type ffprobeResult struct {
	Streams []ffprobeStream `json:"streams"`
	Format  struct {
		FormatName string `json:"format_name"`
		Duration   string `json:"duration"`
		Size       string `json:"size"`
		BitRate    string `json:"bit_rate"`
	} `json:"format"`
}

func parseFFprobeOutput(out []byte) (*VideoMetadata, error) {
	var res ffprobeResult
	if err := json.Unmarshal(out, &res); err != nil {
		return nil, fmt.Errorf("parse ffprobe json: %w", err)
	}
	var video *ffprobeStream
	var audio *ffprobeStream
	for i := range res.Streams {
		s := res.Streams[i]
		switch strings.ToLower(strings.TrimSpace(s.CodecType)) {
		case "video":
			if video == nil {
				video = &s
			}
		case "audio":
			if audio == nil {
				audio = &s
			}
		}
	}
	if video == nil {
		return nil, errors.New("ffprobe: no video stream")
	}
	fps := parseRate(video.AvgFrameRate)
	if fps == 0 {
		fps = parseRate(video.RFrameRate)
	}
	duration := parseDurationSeconds(video.Duration, video.DurationTS, fps)
	if duration == 0 {
		duration = parseFloat(res.Format.Duration)
	}
	meta := &VideoMetadata{
		Codec:           strings.TrimSpace(video.CodecName),
		Container:       strings.TrimSpace(res.Format.FormatName),
		Width:           video.Width,
		Height:          video.Height,
		FPS:             fps,
		DurationSeconds: duration,
	}
	if audio != nil {
		meta.AudioCodec = strings.TrimSpace(audio.CodecName)
		if sr, err := strconv.Atoi(strings.TrimSpace(audio.SampleRate)); err == nil {
			meta.SampleRate = sr
		}
		meta.Channels = audio.Channels
		if meta.DurationSeconds == 0 {
			meta.DurationSeconds = parseDurationSeconds(audio.Duration, audio.DurationTS, 0)
		}
	}
	meta.FormatBitRate = parseInt64(res.Format.BitRate)
	meta.VideoBitRate = parseInt64(video.BitRate)
	if audio != nil {
		meta.AudioBitRate = parseInt64(audio.BitRate)
	}
	return meta, nil
}

func parseRate(rate string) float64 {
	rate = strings.TrimSpace(rate)
	if rate == "" || rate == "0/0" {
		return 0
	}
	if strings.Contains(rate, "/") {
		parts := strings.Split(rate, "/")
		if len(parts) == 2 {
			num, _ := strconv.ParseFloat(parts[0], 64)
			den, _ := strconv.ParseFloat(parts[1], 64)
			if num > 0 && den > 0 {
				return num / den
			}
		}
	}
	v, _ := strconv.ParseFloat(rate, 64)
	return v
}

func parseDurationSeconds(durationStr string, durationTS int64, fps float64) float64 {
	if durationStr != "" {
		if v, err := strconv.ParseFloat(durationStr, 64); err == nil && v > 0 {
			return v
		}
	}
	if durationTS > 0 && fps > 0 {
		return float64(durationTS) / fps
	}
	return 0
}

func parseFloat(raw string) float64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	v, _ := strconv.ParseFloat(raw, 64)
	return v
}

func parseInt64(raw string) int64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	v, _ := strconv.ParseInt(raw, 10, 64)
	return v
}
