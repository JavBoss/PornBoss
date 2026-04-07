//go:build darwin

package util

import (
	"errors"
	"testing"
)

func TestSystemPrefersChineseFallbackUsesAppleLanguages(t *testing.T) {
	original := readDarwinPreference
	t.Cleanup(func() {
		readDarwinPreference = original
	})

	readDarwinPreference = func(key string) (string, error) {
		switch key {
		case "AppleLanguages":
			return "(\n    \"zh-Hans-CN\"\n)", nil
		case "AppleLocale":
			return "", errors.New("should not be needed")
		default:
			t.Fatalf("unexpected preference key %q", key)
			return "", nil
		}
	}

	if !systemPrefersChineseFallback() {
		t.Fatal("systemPrefersChineseFallback() = false, want true")
	}
}
