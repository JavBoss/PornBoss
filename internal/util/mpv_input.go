package util

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const mpvInputConfContent = `a no-osd seek -1 exact
z no-osd seek 1 exact
s no-osd seek -5 exact
x no-osd seek 5 exact
d no-osd seek -30 exact
c no-osd seek 30 exact
f no-osd seek -300 exact
v no-osd seek 300 exact
q add volume -5
w add volume 5
ESC quit
`

const mpvConfigContent = `autofit=70%x70%
geometry=50%:50%
auto-window-resize=no
`

var (
	mpvInputConfOnce sync.Once
	mpvInputConfPath string
	mpvInputConfErr  error

	mpvConfigOnce sync.Once
	mpvConfigPath string
	mpvConfigErr  error
)

func ensureMPVInputConf() (string, error) {
	mpvInputConfOnce.Do(func() {
		mpvInputConfPath, mpvInputConfErr = writeMPVInputConf()
	})
	return mpvInputConfPath, mpvInputConfErr
}

func writeMPVInputConf() (string, error) {
	dir := filepath.Join(os.TempDir(), "pornboss")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create mpv input conf dir: %w", err)
	}

	path := filepath.Join(dir, "mpv-input.conf")
	if err := os.WriteFile(path, []byte(mpvInputConfContent), 0o644); err != nil {
		return "", fmt.Errorf("write mpv input conf: %w", err)
	}

	return path, nil
}

func ensureMPVConfig() (string, error) {
	mpvConfigOnce.Do(func() {
		mpvConfigPath, mpvConfigErr = writeMPVConfig()
	})
	return mpvConfigPath, mpvConfigErr
}

func writeMPVConfig() (string, error) {
	dir := filepath.Join(os.TempDir(), "pornboss")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create mpv config dir: %w", err)
	}

	path := filepath.Join(dir, "mpv.conf")
	if err := os.WriteFile(path, []byte(mpvConfigContent), 0o644); err != nil {
		return "", fmt.Errorf("write mpv config: %w", err)
	}

	return path, nil
}
