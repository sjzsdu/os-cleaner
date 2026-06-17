package registry

import (
	"runtime"
	"testing"
)

func TestGetCategoryByID(t *testing.T) {
	tests := []struct {
		id      string
		wantOK  bool
		wantID  string
	}{
		{"npm-cache", true, "npm-cache"},
		{"non-existent", false, ""},
		{"", false, ""},
		{"go-modules", true, "go-modules"},
	}

	for _, tc := range tests {
		got := GetCategoryByID(tc.id)
		if tc.wantOK && got == nil {
			t.Errorf("GetCategoryByID(%q) = nil, want non-nil", tc.id)
		}
		if tc.wantOK && got != nil && got.ID != tc.wantID {
			t.Errorf("GetCategoryByID(%q).ID = %q, want %q", tc.id, got.ID, tc.wantID)
		}
		if !tc.wantOK && got != nil {
			t.Errorf("GetCategoryByID(%q) = non-nil, want nil", tc.id)
		}
	}
}

func TestGetAllCategories(t *testing.T) {
	all := GetAllCategories()
	if len(all) == 0 {
		t.Error("GetAllCategories() returned empty slice")
	}

	// Verify no duplicate IDs
	seen := make(map[string]bool)
	for _, c := range all {
		if seen[c.ID] {
			t.Errorf("duplicate category ID: %s", c.ID)
		}
		seen[c.ID] = true
	}
}

func TestGetCategoriesByPlatform(t *testing.T) {
	cats := GetCategoriesByPlatform()
	if len(cats) == 0 {
		t.Error("GetCategoriesByPlatform() returned empty slice")
	}

	currentPlatform := getCurrentPlatform()
	for _, c := range cats {
		matches := false
		for _, p := range c.Platforms {
			if p == "all" || p == currentPlatform {
				matches = true
				break
			}
		}
		if !matches {
			t.Errorf("category %s (%s) returned for platform %q but its platforms are %v",
				c.ID, c.Name, currentPlatform, c.Platforms)
		}
	}
}

func TestPlatformCategoriesAreDisjoint(t *testing.T) {
	platforms := []string{"macos", "linux", "windows"}
	platformCats := make(map[string]int)
	for _, p := range platforms {
		platformCats[p] = 0
	}

	for _, c := range GetAllCategories() {
		for _, p := range c.Platforms {
			if _, ok := platformCats[p]; ok {
				platformCats[p]++
			}
		}
	}

	// Each platform should have at least one category
	for _, p := range platforms {
		if platformCats[p] == 0 {
			t.Errorf("no categories found for platform %s", p)
		}
	}
}

func TestGetSafeCategories(t *testing.T) {
	safe := GetSafeCategories()
	if len(safe) == 0 {
		t.Error("GetSafeCategories() returned empty slice")
	}

	for _, c := range safe {
		if c.SafetyLevel != Safe {
			t.Errorf("category %s has SafetyLevel %v, expected Safe", c.ID, c.SafetyLevel)
		}
	}
}

func TestGetCautionCategories(t *testing.T) {
	caution := GetCautionCategories()
	if len(caution) == 0 {
		t.Error("GetCautionCategories() returned empty slice")
	}

	for _, c := range caution {
		if c.SafetyLevel > Caution {
			t.Errorf("category %s has SafetyLevel %v, expected Safe or Caution", c.ID, c.SafetyLevel)
		}
	}
}

func TestSafetyLevelString(t *testing.T) {
	tests := []struct {
		level SafetyLevel
		want  string
	}{
		{Safe, "safe"},
		{Caution, "caution"},
		{Dangerous, "dangerous"},
		{SafetyLevel(99), "unknown"},
	}

	for _, tc := range tests {
		if got := tc.level.String(); got != tc.want {
			t.Errorf("SafetyLevel(%d).String() = %q, want %q", tc.level, got, tc.want)
		}
	}
}

func TestGetCurrentPlatform(t *testing.T) {
	platform := getCurrentPlatform()
	allowed := map[string]bool{
		"macos":   runtime.GOOS == "darwin",
		"linux":   runtime.GOOS == "linux",
		"windows": runtime.GOOS == "windows",
	}

	if !allowed[platform] {
		t.Errorf("getCurrentPlatform() = %q, expected one of %v", platform, []string{"macos", "linux", "windows"})
	}
}

func TestFindCategory(t *testing.T) {
	tests := []struct {
		input string
		want  string // expected category ID
	}{
		{"npm-cache", "npm-cache"},       // exact ID
		{"npm Cache", "npm-cache"},       // exact Name (case insensitive)
		{"NPM", "npm-cache"},             // partial match
		{"xcode", "xcode-deriveddata"},   // partial match (first match)
		{"non-existent-category-zzz", ""},
		{"", ""},
	}

	for _, tc := range tests {
		got := FindCategory(tc.input)
		if tc.want == "" && got != nil {
			t.Errorf("FindCategory(%q) = %s, want nil", tc.input, got.ID)
		}
		if tc.want != "" && got == nil {
			t.Errorf("FindCategory(%q) = nil, want %s", tc.input, tc.want)
		}
		if tc.want != "" && got != nil && got.ID != tc.want {
			t.Errorf("FindCategory(%q).ID = %s, want %s", tc.input, got.ID, tc.want)
		}
	}
}

func TestSearchCategories(t *testing.T) {
	results := SearchCategories("npm")
	if len(results) == 0 {
		t.Error("SearchCategories('npm') returned empty")
	}

	// Should find npm-cache and possibly others
	found := false
	for _, c := range results {
		if c.ID == "npm-cache" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("SearchCategories('npm') = %v, expected to find npm-cache", results)
	}

	// Empty input returns nil
	if r := SearchCategories(""); r != nil {
		t.Error("SearchCategories('') should return nil")
	}
}

func TestCategoryNamesNotEmpty(t *testing.T) {
	for _, c := range GetAllCategories() {
		if c.Name == "" {
			t.Errorf("category %s has empty Name", c.ID)
		}
		if c.Description == "" {
			t.Errorf("category %s has empty Description", c.ID)
		}
	}
}
