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

const configContent = `autofit=70%x70%
geometry=50%:50%
auto-window-resize=no
`

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
	inputConfMu   sync.Mutex
	inputConfPath string

	configOnce sync.Once
	configPath string
	configErr  error
)

func ensureInputConf() (string, error) {
	inputConfMu.Lock()
	defer inputConfMu.Unlock()

	return writeInputConf()
}

func writeInputConf() (string, error) {
	dir := filepath.Join(os.TempDir(), "pornboss")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create mpv input conf dir: %w", err)
	}

	if inputConfPath == "" {
		inputConfPath = filepath.Join(dir, "mpv-input.conf")
	}

	content, err := buildInputConfContent()
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(inputConfPath, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write mpv input conf: %w", err)
	}

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
	configOnce.Do(func() {
		configPath, configErr = writeConfig()
	})
	return configPath, configErr
}

func writeConfig() (string, error) {
	dir := filepath.Join(os.TempDir(), "pornboss")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create mpv config dir: %w", err)
	}

	path := filepath.Join(dir, "mpv.conf")
	if err := os.WriteFile(path, []byte(configContent), 0o644); err != nil {
		return "", fmt.Errorf("write mpv config: %w", err)
	}

	return path, nil
}
