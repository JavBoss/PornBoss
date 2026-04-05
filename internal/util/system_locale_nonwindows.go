//go:build !windows

package util

func systemPrefersChineseFallback() bool {
	return false
}
