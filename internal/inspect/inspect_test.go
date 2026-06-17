package inspect

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/juzhongsun/os-cleaner/internal/registry"
)

func TestParseSize(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"1K", 1024},
		{"1M", 1024 * 1024},
		{"1G", 1024 * 1024 * 1024},
		{"2.5M", int64(2.5 * 1024 * 1024)},
		{"512", 512},
		{"", 0},
		// edge cases
		{"0", 0},
		{"0G", 0},
		{"0K", 0},
		{"0M", 0},
		// parseSize is a best-effort parser; negative values pass through
		{"abc", 0},
		{"  1G  ", 1024 * 1024 * 1024},
		{"1.5G", int64(1.5 * 1024 * 1024 * 1024)},
		{"10", 10},
		{"100K", 100 * 1024},
		{"1024", 1024},
		{"1.0M", int64(1.0 * 1024 * 1024)},
	}

	for _, tc := range tests {
		if got := parseSize(tc.input); got != tc.want {
			t.Errorf("parseSize(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestFormatCount(t *testing.T) {
	// formatCount returns empty string for < 1000, and ANSI-wrapped string for >= 1000
	if got := formatCount(0); got != "" {
		t.Errorf("formatCount(0) = %q, want empty", got)
	}
	if got := formatCount(500); got != "" {
		t.Errorf("formatCount(500) = %q, want empty", got)
	}
	if got := formatCount(999); got != "" {
		t.Errorf("formatCount(999) = %q, want empty", got)
	}
	// Threshold is > 1000 (strictly greater)
	if got := formatCount(1000); got != "" {
		t.Errorf("formatCount(1000) = %q, want empty (threshold is > 1000)", got)
	}
	// For > 1000, it returns a non-empty string wrapped in ANSI codes
	if got := formatCount(1001); got == "" {
		t.Error("formatCount(1001) should be non-empty")
	}
	if got := formatCount(1000000); got == "" {
		t.Error("formatCount(1000000) should be non-empty")
	}
}

func TestFormatWithCommas(t *testing.T) {
	tests := []struct {
		n    int64
		want string
	}{
		{0, " items"},
		{1, "1 items"},
		{10, "10 items"},
		{100, "100 items"},
		{999, "999 items"},
		{1000, "1,000 items"},
		{1001, "1,001 items"},
		{10000, "10,000 items"},
		{100000, "100,000 items"},
		{1234567, "1,234,567 items"},
		{12345678, "12,345,678 items"},
		{123456789, "123,456,789 items"},
		{1000000000, "1,000,000,000 items"},
	}

	for _, tc := range tests {
		if got := formatWithCommas(tc.n); got != tc.want {
			t.Errorf("formatWithCommas(%d) = %q, want %q", tc.n, got, tc.want)
		}
	}
}

func TestHomeDirWithSlash(t *testing.T) {
	got := homeDirWithSlash()
	if len(got) == 0 {
		t.Fatal("homeDirWithSlash() returned empty string")
	}
	if got[len(got)-1] != '/' {
		t.Errorf("homeDirWithSlash() = %q, want trailing slash", got)
	}
}

func TestIsNPMCache(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{"npm-cache", true},
		{"npm-cache-windows", true},
		{"go-modules", false},
		{"docker", false},
		{"", false},
		{"npm", true},  // contains "npm"
		{"some-npm-thing", true},
	}
	for _, tc := range tests {
		cat := registry.CacheCategory{ID: tc.id}
		if got := isNPMCache(cat); got != tc.want {
			t.Errorf("isNPMCache(%q) = %v, want %v", tc.id, got, tc.want)
		}
	}
}

func TestIsGoModules(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{"go-modules", true},
		{"go-build-cache", false},
		{"npm-cache", false},
		{"", false},
	}
	for _, tc := range tests {
		cat := registry.CacheCategory{ID: tc.id}
		if got := isGoModules(cat); got != tc.want {
			t.Errorf("isGoModules(%q) = %v, want %v", tc.id, got, tc.want)
		}
	}
}

func TestIsPythonPyenv(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{"python-pyenv", true},
		{"python-pip-cache", false},
		{"", false},
		{"pyenv", false},
	}
	for _, tc := range tests {
		cat := registry.CacheCategory{ID: tc.id}
		if got := isPythonPyenv(cat); got != tc.want {
			t.Errorf("isPythonPyenv(%q) = %v, want %v", tc.id, got, tc.want)
		}
	}
}

func TestIsPythonPip(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{"python-pip-cache", true},
		{"python-pyenv", false},
		{"pip-cache-windows", false},
		{"", false},
	}
	for _, tc := range tests {
		cat := registry.CacheCategory{ID: tc.id}
		if got := isPythonPip(cat); got != tc.want {
			t.Errorf("isPythonPip(%q) = %v, want %v", tc.id, got, tc.want)
		}
	}
}

func TestIsCargo(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{"cargo-registry", true},
		{"cargo-git", true},
		{"npm-cache", false},
		{"", false},
	}
	for _, tc := range tests {
		cat := registry.CacheCategory{ID: tc.id}
		if got := isCargo(cat); got != tc.want {
			t.Errorf("isCargo(%q) = %v, want %v", tc.id, got, tc.want)
		}
	}
}

func TestIsMaven(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{"maven-repository", true},
		{"gradle-cache", false},
		{"", false},
	}
	for _, tc := range tests {
		cat := registry.CacheCategory{ID: tc.id}
		if got := isMaven(cat); got != tc.want {
			t.Errorf("isMaven(%q) = %v, want %v", tc.id, got, tc.want)
		}
	}
}

func TestIsGradle(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{"gradle-cache", true},
		{"maven-repository", false},
		{"", false},
	}
	for _, tc := range tests {
		cat := registry.CacheCategory{ID: tc.id}
		if got := isGradle(cat); got != tc.want {
			t.Errorf("isGradle(%q) = %v, want %v", tc.id, got, tc.want)
		}
	}
}

func TestIsHomebrew(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{"homebrew-cache", true},
		{"apt-cache", false},
		{"", false},
	}
	for _, tc := range tests {
		cat := registry.CacheCategory{ID: tc.id}
		if got := isHomebrew(cat); got != tc.want {
			t.Errorf("isHomebrew(%q) = %v, want %v", tc.id, got, tc.want)
		}
	}
}

func TestIsXcode(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{"xcode-deriveddata", true},
		{"xcode-simulator", true},
		{"xcode-archives", true},
		{"xcode-devicesupport", true},
		{"homebrew-cache", false},
		{"", false},
	}
	for _, tc := range tests {
		cat := registry.CacheCategory{ID: tc.id}
		if got := isXcode(cat); got != tc.want {
			t.Errorf("isXcode(%q) = %v, want %v", tc.id, got, tc.want)
		}
	}
}

func TestGetDirSizeAndCount(t *testing.T) {
	dir := t.TempDir()

	// Empty dir
	size, count := getDirSizeAndCount(dir)
	if size != 0 {
		t.Errorf("empty dir size = %d, want 0", size)
	}
	if count != 0 {
		t.Errorf("empty dir count = %d, want 0", count)
	}

	// Dir with files
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("world"), 0644); err != nil {
		t.Fatal(err)
	}

	size, count = getDirSizeAndCount(dir)
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
	if size <= 0 {
		t.Errorf("size = %d, want > 0", size)
	}

	// Nested dir
	subdir := filepath.Join(dir, "sub")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subdir, "c.txt"), []byte("nested"), 0644); err != nil {
		t.Fatal(err)
	}

	size, count = getDirSizeAndCount(dir)
	if count != 3 {
		t.Errorf("with subdir count = %d, want 3", count)
	}

	// Nonexistent dir
	size, count = getDirSizeAndCount("/nonexistent/path")
	if size != 0 || count != 0 {
		t.Errorf("nonexistent: size=%d, count=%d, want 0,0", size, count)
	}
}

func TestDoInspectByPath(t *testing.T) {
	dir := t.TempDir()

	// Inspect nonexistent path should not error
	err := doInspectByPath(dir+"-nonexistent", 10)
	if err != nil {
		t.Errorf("inspect nonexistent path: %v", err)
	}

	// Inspect real path (existing dir)
	if err := os.WriteFile(filepath.Join(dir, "test.txt"), []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	err = doInspectByPath(dir, 10)
	if err != nil {
		t.Errorf("inspect real path: %v", err)
	}
}

func TestRun(t *testing.T) {
	// Run with empty category should error
	err := Run(InspectOptions{})
	if err == nil {
		t.Error("Run() with empty options should error")
	}

	// Run with --all on nonexistent categories should not error
	err = Run(InspectOptions{All: true})
	if err != nil {
		t.Errorf("Run(All=true): %v", err)
	}

	// Run with nonexistent category should not error (falls through to path inspect)
	err = Run(InspectOptions{Category: "__nonexistent_category_xyz__"})
	if err != nil {
		t.Errorf("Run(nonexistent category): %v", err)
	}
}
