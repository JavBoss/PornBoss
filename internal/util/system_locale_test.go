package util

import "testing"

func TestIsChineseLocaleValue(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{value: "", want: false},
		{value: "en_US.UTF-8", want: false},
		{value: "zh_CN.UTF-8", want: true},
		{value: "zh-Hans-CN", want: true},
		{value: "English:Chinese", want: true},
		{value: "ja_JP:en_US", want: false},
	}

	for _, tt := range tests {
		if got := isChineseLocaleValue(tt.value); got != tt.want {
			t.Fatalf("isChineseLocaleValue(%q)=%v want %v", tt.value, got, tt.want)
		}
	}
}
