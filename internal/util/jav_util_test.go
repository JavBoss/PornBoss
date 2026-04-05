package util_test

import (
	"testing"

	"pornboss/internal/util"
)

func TestExtractCodeFromName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "1024核工厂-ABP-356.HD.mp4",
			expected: "ABP-356",
		},
		{
			input:    "ABP-178.avi",
			expected: "ABP-178",
		},
		{
			input:    "ABP-888C.mp4",
			expected: "ABP-888",
		},
		{
			input:    "abp-782_1.mp4",
			expected: "ABP-782",
		},
		{
			input:    "AGEMIX-080.avi",
			expected: "AGEMIX-080",
		},
		{
			input:    "AMBI027a.avi",
			expected: "AMBI-027",
		},
		{
			input:    "[44x.me]AOZ-132-C.mp4",
			expected: "AOZ-132",
		},
		{
			input:    "IBW-938z.mp4",
			expected: "IBW-938",
		},
		{
			input:    "BANK-002 S .mp4",
			expected: "BANK-002",
		},
		{
			input:    "dv-1530_0.avi",
			expected: "DV-1530",
		},
		{
			input:    "dv-1448.mp4",
			expected: "DV-1448",
		},
		{
			input:    "SDNM256.mp4",
			expected: "SDNM-256",
		},
		{
			input:    "PKPD066C.mp4",
			expected: "PKPD-066",
		},
		{
			input:    "miaa00068.mp4",
			expected: "MIAA-068",
		},
		{
			input:    "ovg00129.mp4",
			expected: "OVG-129",
		},
		{
			input:    "[XJ0610.com]IBW-497z-720p.mp4",
			expected: "IBW-497Z",
		},
		{
			input:    "IBW572Z.mp4",
			expected: "IBW-572Z",
		},
	}

	for _, tt := range tests {
		got := util.ExtractCodeFromName(tt.input)
		if !contains(got, tt.expected) {
			t.Fatalf("ExtractCodeFromName(%q) missing expected code %q in %#v", tt.input, tt.expected, got)
		}
	}
}

func contains(list []string, needle string) bool {
	for _, v := range list {
		if v == needle {
			return true
		}
	}
	return false
}
