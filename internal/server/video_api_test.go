package server

import "testing"

func TestIsScreenshotImageName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{name: "00-12-34.567.jpg", want: true},
		{name: "example.PNG", want: true},
		{name: "../example.jpg", want: false},
		{name: "nested/example.jpg", want: false},
		{name: "example.txt", want: false},
		{name: "", want: false},
	}

	for _, tt := range tests {
		if got := isScreenshotImageName(tt.name); got != tt.want {
			t.Fatalf("isScreenshotImageName(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}
