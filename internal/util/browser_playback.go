package util

import (
	"path/filepath"
	"strings"
)

const (
	BrowserPlaybackModeDirect = "direct"
	BrowserPlaybackModeFLV    = "flv"
	BrowserPlaybackModeHLS    = "hls"
)

func DetermineBrowserPlayback(path string, meta *VideoMetadata) (string, string) {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".flv" {
		return BrowserPlaybackModeFLV, "video/x-flv"
	}

	if mimeType := DetectBrowserDirectMime(path, meta); mimeType != "" {
		return BrowserPlaybackModeDirect, mimeType
	}

	return BrowserPlaybackModeHLS, "application/vnd.apple.mpegurl"
}

func DetectBrowserDirectMime(path string, meta *VideoMetadata) string {
	if meta == nil {
		switch normalizeContainer("", path) {
		case "mp4":
			return "video/mp4"
		case "webm":
			return "video/webm"
		default:
			return ""
		}
	}

	container := normalizeContainer(meta.Container, path)
	videoCodec := normalizeVideoCodec(meta.Codec)
	audioCodec := normalizeAudioCodec(meta.AudioCodec)

	switch container {
	case "mp4":
		if videoCodec == "h264" && isOneOf(audioCodec, "", "aac", "mp3", "opus") {
			return "video/mp4"
		}
	case "webm":
		if isOneOf(videoCodec, "vp8", "vp9") && isOneOf(audioCodec, "", "vorbis", "opus") {
			return "video/webm"
		}
	}

	return ""
}

func normalizeContainer(raw, path string) string {
	for _, part := range strings.Split(strings.ToLower(strings.TrimSpace(raw)), ",") {
		switch strings.TrimSpace(part) {
		case "mov", "mp4", "m4a", "3gp", "3g2", "mj2", "ismv":
			return "mp4"
		case "webm":
			return "webm"
		case "matroska":
			return "matroska"
		case "avi":
			return "avi"
		case "mpegts":
			return "mpegts"
		case "flv":
			return "flv"
		case "asf":
			return "wmv"
		}
	}

	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp4", ".m4v", ".mov":
		return "mp4"
	case ".webm":
		return "webm"
	case ".mkv":
		return "matroska"
	case ".avi":
		return "avi"
	case ".ts", ".mts", ".m2ts":
		return "mpegts"
	case ".flv":
		return "flv"
	default:
		return ""
	}
}

func normalizeVideoCodec(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "h264", "avc1":
		return "h264"
	case "h265", "hevc":
		return "h265"
	case "vp8":
		return "vp8"
	case "vp9":
		return "vp9"
	default:
		return ""
	}
}

func normalizeAudioCodec(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "":
		return ""
	case "aac":
		return "aac"
	case "mp3":
		return "mp3"
	case "opus":
		return "opus"
	case "vorbis":
		return "vorbis"
	default:
		return ""
	}
}

func isOneOf(value string, candidates ...string) bool {
	for _, candidate := range candidates {
		if value == candidate {
			return true
		}
	}
	return false
}
