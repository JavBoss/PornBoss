package mpv

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"pornboss/internal/common"
	"pornboss/internal/common/logging"
	dbpkg "pornboss/internal/db"
)

const playerHotkeysConfigKey = "player_hotkeys"
const playerWindowSizeConfigKey = "player_window_size"
const playerWindowWidthConfigKey = "player_window_width"
const playerWindowHeightConfigKey = "player_window_height"
const playerWindowUseAutofitConfigKey = "player_window_use_autofit"
const playerVolumeConfigKey = "player_volume"
const playerOntopConfigKey = "player_ontop"

const (
	defaultWindowWidth  = 70
	defaultWindowHeight = 70
	defaultVolume       = 70
	defaultOntop        = true
)

type hotkeyConfig struct {
	Key    string  `json:"key"`
	Action string  `json:"action"`
	Amount float64 `json:"amount"`
}

var defaultHotkeys = []hotkeyConfig{
	{Key: "a", Action: "seek", Amount: -1},
	{Key: "z", Action: "seek", Amount: 1},
	{Key: "s", Action: "seek", Amount: -5},
	{Key: "x", Action: "seek", Amount: 5},
	{Key: "d", Action: "seek", Amount: -30},
	{Key: "c", Action: "seek", Amount: 30},
	{Key: "f", Action: "seek", Amount: -300},
	{Key: "v", Action: "seek", Amount: 300},
	{Key: "q", Action: "volume", Amount: -5},
	{Key: "w", Action: "volume", Amount: 5},
}

var (
	inputConfMu      sync.Mutex
	inputConfPath    string
	inputConfContent string
	inputConfReady   bool

	configMu      sync.Mutex
	configPath    string
	configContent string
	configReady   bool
)

func InvalidateHotkeysCache() {
	inputConfMu.Lock()
	defer inputConfMu.Unlock()

	inputConfContent = ""
	inputConfReady = false
}

func InvalidatePlayerConfigCache() {
	configMu.Lock()
	defer configMu.Unlock()

	configContent = ""
	configReady = false
}

func ensureInputConf() (string, error) {
	inputConfMu.Lock()
	defer inputConfMu.Unlock()

	dir := filepath.Join(os.TempDir(), "pornboss")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create mpv input conf dir: %w", err)
	}

	if inputConfPath == "" {
		inputConfPath = filepath.Join(dir, "mpv-input.conf")
	}

	if inputConfReady {
		if _, err := os.Stat(inputConfPath); err == nil {
			return inputConfPath, nil
		}
		if err := os.WriteFile(inputConfPath, []byte(inputConfContent), 0o644); err != nil {
			return "", fmt.Errorf("restore mpv input conf: %w", err)
		}
		return inputConfPath, nil
	}

	return writeInputConf()
}

func writeInputConf() (string, error) {
	content, err := buildInputConfContent()
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(inputConfPath, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write mpv input conf: %w", err)
	}

	inputConfContent = content
	inputConfReady = true

	return inputConfPath, nil
}

func buildInputConfContent() (string, error) {
	hotkeys, err := loadConfiguredHotkeys()
	if err != nil {
		return "", err
	}

	var lines []string
	for _, item := range hotkeys {
		keyName, ok := keyName(item.Key)
		if !ok {
			continue
		}

		switch item.Action {
		case "seek":
			lines = append(lines, fmt.Sprintf("%s no-osd seek %s exact", keyName, formatAmount(item.Amount)))
		case "volume":
			lines = append(lines, fmt.Sprintf("%s add volume %s", keyName, formatAmount(item.Amount)))
		}
	}

	lines = append(lines, "ESC quit")
	return strings.Join(lines, "\n") + "\n", nil
}

func loadConfiguredHotkeys() ([]hotkeyConfig, error) {
	if common.DB == nil {
		return cloneDefaultHotkeys(), nil
	}

	cfg, err := dbpkg.ListConfig(context.Background())
	if err != nil {
		logging.Error("list player_hotkeys config failed, using defaults: %v", err)
		return cloneDefaultHotkeys(), nil
	}

	raw := strings.TrimSpace(cfg[playerHotkeysConfigKey])
	if raw == "" {
		return cloneDefaultHotkeys(), nil
	}

	var parsed []hotkeyConfig
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		logging.Error("parse player_hotkeys config failed, using defaults: %v", err)
		return cloneDefaultHotkeys(), nil
	}

	return normalizeHotkeys(parsed), nil
}

func cloneDefaultHotkeys() []hotkeyConfig {
	items := make([]hotkeyConfig, len(defaultHotkeys))
	copy(items, defaultHotkeys)
	return items
}

func normalizeHotkeys(items []hotkeyConfig) []hotkeyConfig {
	if len(items) == 0 {
		return nil
	}

	normalized := make([]hotkeyConfig, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		key := normalizeHotkeyKey(item.Key)
		if key == "" || key == " " || strings.EqualFold(key, "Escape") {
			continue
		}

		action := strings.ToLower(strings.TrimSpace(item.Action))
		if action != "seek" && action != "volume" {
			continue
		}
		if item.Amount == 0 {
			continue
		}
		if action == "volume" && (item.Amount < -100 || item.Amount > 100) {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		if _, ok := keyName(key); !ok {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, hotkeyConfig{
			Key:    key,
			Action: action,
			Amount: item.Amount,
		})
	}

	return normalized
}

func normalizeHotkeyKey(raw string) string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return ""
	}
	if len(text) == 1 {
		return strings.ToLower(text)
	}
	if text == "Esc" {
		return "Escape"
	}
	if text == "Spacebar" {
		return " "
	}
	return text
}

func keyName(key string) (string, bool) {
	normalized := normalizeHotkeyKey(key)
	if normalized == "" || normalized == " " || strings.EqualFold(normalized, "Escape") {
		return "", false
	}
	if len(normalized) == 1 {
		return normalized, true
	}

	switch normalized {
	case "ArrowLeft":
		return "LEFT", true
	case "ArrowRight":
		return "RIGHT", true
	case "ArrowUp":
		return "UP", true
	case "ArrowDown":
		return "DOWN", true
	case "Enter":
		return "ENTER", true
	case "Backspace":
		return "BS", true
	case "Delete":
		return "DEL", true
	case "Insert":
		return "INS", true
	case "Home":
		return "HOME", true
	case "End":
		return "END", true
	case "PageUp":
		return "PGUP", true
	case "PageDown":
		return "PGDWN", true
	default:
		if strings.HasPrefix(normalized, "F") {
			n, err := strconv.Atoi(normalized[1:])
			if err == nil && n >= 1 && n <= 12 {
				return normalized, true
			}
		}
		return "", false
	}
}

func formatAmount(amount float64) string {
	return strconv.FormatFloat(amount, 'f', -1, 64)
}

func ensureConfig() (string, error) {
	configMu.Lock()
	defer configMu.Unlock()

	dir := filepath.Join(os.TempDir(), "pornboss")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create mpv config dir: %w", err)
	}

	if configPath == "" {
		configPath = filepath.Join(dir, "mpv.conf")
	}

	if configReady {
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
		if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
			return "", fmt.Errorf("restore mpv config: %w", err)
		}
		return configPath, nil
	}

	return writeConfig()
}

func writeConfig() (string, error) {
	content, err := buildConfigContent()
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write mpv config: %w", err)
	}

	configContent = content
	configReady = true

	return configPath, nil
}

func buildConfigContent() (string, error) {
	windowWidth, windowHeight, useAutofit, volume, ontop, err := loadConfiguredPlayerBaseSettings()
	if err != nil {
		return "", err
	}

	lines := []string{
		"keep-open=yes",
		fmt.Sprintf("ontop=%s", mpvBool(ontop)),
	}
	if useAutofit {
		lines = append(lines, fmt.Sprintf("autofit=%d%%x%d%%", windowWidth, windowHeight))
	} else {
		lines = append(lines,
			"auto-window-resize=no",
			"geometry="+centeredWindowGeometry(windowWidth, windowHeight),
		)
	}
	lines = append(lines, fmt.Sprintf("volume=%d", volume))

	return strings.Join(lines, "\n") + "\n", nil
}

func centeredWindowGeometry(width, height int) string {
	return fmt.Sprintf("%d%%x%d%%+50%%+50%%", width, height)
}

func loadConfiguredPlayerBaseSettings() (int, int, bool, int, bool, error) {
	if common.DB == nil {
		return defaultWindowWidth, defaultWindowHeight, false, defaultVolume, defaultOntop, nil
	}

	cfg, err := dbpkg.ListConfig(context.Background())
	if err != nil {
		logging.Error("list player base config failed, using defaults: %v", err)
		return defaultWindowWidth, defaultWindowHeight, false, defaultVolume, defaultOntop, nil
	}

	windowWidth := defaultWindowWidth
	windowHeight := defaultWindowHeight
	if raw := strings.TrimSpace(cfg[playerWindowSizeConfigKey]); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed >= 10 && parsed <= 100 {
			windowWidth = parsed
			windowHeight = parsed
		}
	}
	if raw := strings.TrimSpace(cfg[playerWindowWidthConfigKey]); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed >= 10 && parsed <= 100 {
			windowWidth = parsed
		}
	}
	if raw := strings.TrimSpace(cfg[playerWindowHeightConfigKey]); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed >= 10 && parsed <= 100 {
			windowHeight = parsed
		}
	}

	useAutofit := false
	if raw := strings.TrimSpace(cfg[playerWindowUseAutofitConfigKey]); raw != "" {
		switch strings.ToLower(raw) {
		case "0", "false", "no", "off":
			useAutofit = false
		case "1", "true", "yes", "on":
			useAutofit = true
		}
	}

	volume := defaultVolume
	if raw := strings.TrimSpace(cfg[playerVolumeConfigKey]); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed >= 0 && parsed <= 130 {
			volume = parsed
		}
	}

	ontop := defaultOntop
	if raw := strings.TrimSpace(cfg[playerOntopConfigKey]); raw != "" {
		switch strings.ToLower(raw) {
		case "0", "false", "no", "off":
			ontop = false
		case "1", "true", "yes", "on":
			ontop = true
		}
	}

	return windowWidth, windowHeight, useAutofit, volume, ontop, nil
}

func mpvBool(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
