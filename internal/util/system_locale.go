package util

import (
	"os"
	"strings"
	"sync"
)

var (
	systemPrefersChineseOnce sync.Once
	systemPrefersChinese     bool
)

// SystemPrefersChinese reports whether the current OS/UI locale should be treated as Chinese.
func SystemPrefersChinese() bool {
	systemPrefersChineseOnce.Do(func() {
		for _, key := range []string{"LANGUAGE", "LC_ALL", "LC_MESSAGES", "LANG", "AppleLocale"} {
			if isChineseLocaleValue(os.Getenv(key)) {
				systemPrefersChinese = true
				return
			}
		}
		systemPrefersChinese = systemPrefersChineseFallback()
	})
	return systemPrefersChinese
}

func isChineseLocaleValue(value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return false
	}
	for _, part := range strings.Split(value, ":") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, "zh") || strings.Contains(part, "_zh") || strings.Contains(part, "-zh") || strings.Contains(part, "chinese") {
			return true
		}
	}
	return false
}
