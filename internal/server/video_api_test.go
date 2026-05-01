package server

import "testing"

func TestIsScreenshotImageName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{name: "mpv_00-12-34.567.jpg", want: true},
		{name: "mpv_example.PNG", want: true},
		{name: "00-12-34.567.jpg", want: false},
		{name: "../mpv_example.jpg", want: false},
		{name: "nested/mpv_example.jpg", want: false},
		{name: "mpv_example.txt", want: false},
		{name: "", want: false},
	}

	for _, tt := range tests {
		if got := isScreenshotImageName(tt.name); got != tt.want {
			t.Fatalf("isScreenshotImageName(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}
