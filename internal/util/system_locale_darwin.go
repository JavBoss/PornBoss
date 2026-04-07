//go:build darwin

package util

import (
	"os/exec"
	"strings"
)

var readDarwinPreference = readDarwinPreferenceFromDefaults

func systemPrefersChineseFallback() bool {
	for _, key := range []string{"AppleLanguages", "AppleLocale"} {
		value, err := readDarwinPreference(key)
		if err != nil {
			continue
		}
		if isChineseLocaleValue(value) {
			return true
		}
	}
	return false
}

func readDarwinPreferenceFromDefaults(key string) (string, error) {
	output, err := exec.Command("defaults", "read", "-g", key).CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
