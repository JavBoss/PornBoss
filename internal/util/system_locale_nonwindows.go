//go:build !windows && !darwin

package util

func systemPrefersChineseFallback() bool {
	return false
}
