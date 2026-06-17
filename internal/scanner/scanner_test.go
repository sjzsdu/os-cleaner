package scanner

import (
	"testing"
)

func TestFilesDisplay(t *testing.T) {
	tests := []struct {
		count int64
		want  string
	}{
		{0, "-"},
		{1, "1"},
		{100, "100"},
		{12345, "12345"},
	}

	for _, tc := range tests {
		r := &ScanResult{FileCount: tc.count}
		if got := r.filesDisplay(); got != tc.want {
			t.Errorf("filesDisplay(%d) = %q, want %q", tc.count, got, tc.want)
		}
	}
}

func TestFilesDisplayEmpty(t *testing.T) {
	r := &ScanResult{FileCount: 0}
	if r.filesDisplay() != "-" {
		t.Errorf("filesDisplay() with 0 count should return '-', got %q", r.filesDisplay())
	}
}

