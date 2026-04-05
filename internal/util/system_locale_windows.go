//go:build windows

package util

import "golang.org/x/sys/windows"

func systemPrefersChineseFallback() bool {
	languages, err := windows.GetUserPreferredUILanguages(windows.MUI_LANGUAGE_NAME)
	if err != nil {
		return false
	}
	for _, language := range languages {
		if isChineseLocaleValue(language) {
			return true
		}
	}
	return false
}
