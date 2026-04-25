package mpv

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"pornboss/internal/common/logging"
)

// PlayVideo launches mpv to play the given file path.
func PlayVideo(path string) error {
	cmd, err := buildCommand(path)
	if err != nil {
		return err
	}
	if err := startCommand(cmd, "play video"); err != nil {
		return fmt.Errorf("play video: %w", err)
	}
	return nil
}

func buildCommand(path string) (*exec.Cmd, error) {
	mpvPath, err := ResolvePath()
	if err != nil {
		return nil, err
	}
	inputConfPath, err := ensureInputConf()
	if err != nil {
		return nil, err
	}
	mpvConfigPath, err := ensureConfig()
	if err != nil {
		return nil, err
	}
	args := make([]string, 0, 8)
	if runtime.GOOS == "linux" && os.Getenv("PORNBOSS_BUILD_MODE") != "release" {
		args = append(args, "--vo=x11")
	}
	args = append(args, "--include="+mpvConfigPath)
	args = append(args,
		"--script-opt=osc-visibility=always",
		"--script-opt=osc-layout=bottombar",
		"--script-opt=osc-boxvideo=yes",
	)
	args = append(args, "--input-conf="+inputConfPath)
	args = append(args, "--", path)
	return exec.Command(mpvPath, args...), nil
}

func startCommand(cmd *exec.Cmd, label string) error {
	logging.Info("%s command: %v", label, cmd.Args)
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		if err := cmd.Wait(); err != nil {
			logging.Error("%s command exited with error: %v", label, err)
		}
	}()
	return nil
}
