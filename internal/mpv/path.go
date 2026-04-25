package mpv

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var (
	mpvOnce sync.Once
	mpvPath string
	mpvErr  error
)

// ResolvePath returns the path to the mpv binary by checking:
//  1. MPV_PATH env var
//  2. internal/bin/mpv(.exe) relative to current working directory
//  3. internal/bin/mpv/mpv(.exe) relative to current working directory
//  4. internal/bin/mpv(.exe) relative to the running executable
//  5. internal/bin/mpv/mpv(.exe) relative to the running executable
//  6. bundled macOS app locations under internal/bin
//  7. mpv(.exe) in PATH
//
// The result is cached after the first resolution attempt.
func ResolvePath() (string, error) {
	mpvOnce.Do(func() {
		mpvPath, mpvErr = findPath()
	})
	return mpvPath, mpvErr
}

func findPath() (string, error) {
	var candidates []string

	if env := strings.TrimSpace(os.Getenv("MPV_PATH")); env != "" {
		candidates = append(candidates, env)
	}

	binName := "mpv"
	if runtime.GOOS == "windows" {
		binName = "mpv.exe"
	}

	if wd, err := os.Getwd(); err == nil {
		candidates = appendCandidates(candidates, wd, binName)
	}

	if execPath, err := os.Executable(); err == nil {
		candidates = appendCandidates(candidates, filepath.Dir(execPath), binName)
	}

	candidates = append(candidates, binName)
	if binName != "mpv" {
		candidates = append(candidates, "mpv")
	}

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if resolved, err := exec.LookPath(candidate); err == nil {
			return resolved, nil
		}
	}

	return "", errors.New("mpv not found; set MPV_PATH or place bundled mpv at internal/bin/mpv")
}

func appendCandidates(candidates []string, baseDir, binName string) []string {
	candidates = append(
		candidates,
		filepath.Join(baseDir, "internal", "bin", binName),
		filepath.Join(baseDir, "internal", "bin", "mpv", binName),
	)
	if runtime.GOOS == "darwin" {
		candidates = append(
			candidates,
			filepath.Join(baseDir, "internal", "bin", "mpv", "mpv.app", "Contents", "MacOS", "mpv"),
			filepath.Join(baseDir, "internal", "bin", "mpv.app", "Contents", "MacOS", "mpv"),
		)
	}
	return candidates
}
