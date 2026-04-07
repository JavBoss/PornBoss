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
		for _, key := range []string{"LANGUAGE", "LC_ALL", "LC_MESSAGES", "LANG", "AppleLanguages", "AppleLocale"} {
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

	parts := strings.FieldsFunc(value, func(r rune) bool {
		switch r {
		case ':', ';', ',', '\n', '\r', '\t', ' ', '(', ')', '[', ']', '{', '}', '"', '\'':
			return true
		default:
			return false
		}
	})

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if idx := strings.IndexAny(part, ".@"); idx >= 0 {
			part = part[:idx]
		}
		if part == "zh" || strings.HasPrefix(part, "zh_") || strings.HasPrefix(part, "zh-") || strings.Contains(part, "chinese") {
			return true
		}
	}
	return false
}
