package registry

import (
	"runtime"
)

// SafetyLevel represents the safety level of cleaning a cache
type SafetyLevel int

const (
	// Safe can be cleaned without any risk
	Safe SafetyLevel = iota
	// Caution can be cleaned but may require rebuilds
	Caution
	// Dangerous should not be cleaned automatically
	Dangerous
)

func (s SafetyLevel) String() string {
	switch s {
	case Safe:
		return "safe"
	case Caution:
		return "caution"
	case Dangerous:
		return "dangerous"
	default:
		return "unknown"
	}
}

// CacheCategory defines a cleanable cache category
type CacheCategory struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Description    string      `json:"description"`
	Platforms      []string    `json:"platforms"` // "macos", "linux", "all"
	SafetyLevel    SafetyLevel `json:"safety_level"`
	Paths          []PathRule  `json:"paths"`
	CleanCmd       string      `json:"clean_cmd,omitempty"`            // Optional cleanup command
	SuggestMinSize int64       `json:"suggest_min_size,omitempty"`     // Minimum size (bytes) to show in default scan; hides trivial caches
}

// PathRule defines a path pattern for scanning
type PathRule struct {
	Path    string `json:"path"`    // Can contain ~/ or $HOME
	Pattern string `json:"pattern"` // Optional glob pattern
}

// GetAllCategories returns all registered cache categories
func GetAllCategories() []CacheCategory {
	return categories
}

// GetCategoryByID returns a category by its ID
func GetCategoryByID(id string) *CacheCategory {
	for i := range categories {
		if categories[i].ID == id {
			return &categories[i]
		}
	}
	return nil
}

// GetCategoriesByPlatform returns categories for the current platform
func GetCategoriesByPlatform() []CacheCategory {
	var result []CacheCategory
	currentPlatform := getCurrentPlatform()

	for _, cat := range categories {
		for _, p := range cat.Platforms {
			if p == "all" || p == currentPlatform {
				result = append(result, cat)
				break
			}
		}
	}
	return result
}

// GetSafeCategories returns only safe categories
func GetSafeCategories() []CacheCategory {
	var result []CacheCategory
	platformCategories := GetCategoriesByPlatform()

	for _, cat := range platformCategories {
		if cat.SafetyLevel == Safe {
			result = append(result, cat)
		}
	}
	return result
}

// GetCautionCategories returns safe + caution categories
func GetCautionCategories() []CacheCategory {
	var result []CacheCategory
	platformCategories := GetCategoriesByPlatform()

	for _, cat := range platformCategories {
		if cat.SafetyLevel <= Caution {
			result = append(result, cat)
		}
	}
	return result
}

// FindCategory finds a category by ID, name, or partial match.
// Priority: exact ID → exact Name (case insensitive) → contains ID or Name (case insensitive).
func FindCategory(input string) *CacheCategory {
	if input == "" {
		return nil
	}
	inputLower := toLower(input)

	// Exact ID match
	if c := GetCategoryByID(input); c != nil {
		return c
	}

	// Exact Name match (case insensitive)
	for i := range categories {
		if toLower(categories[i].Name) == inputLower {
			return &categories[i]
		}
	}

	// Partial match on ID or Name (case insensitive)
	for i := range categories {
		if containsLower(categories[i].ID, inputLower) || containsLower(categories[i].Name, inputLower) {
			return &categories[i]
		}
	}

	return nil
}

// SearchCategories finds all categories whose ID or Name contains the input (case insensitive).
func SearchCategories(input string) []CacheCategory {
	if input == "" {
		return nil
	}
	inputLower := toLower(input)

	var result []CacheCategory
	for _, c := range categories {
		if containsLower(c.ID, inputLower) || containsLower(c.Name, inputLower) {
			result = append(result, c)
		}
	}
	return result
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		b[i] = c
	}
	return string(b)
}

func containsLower(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[:len(substr)] == substr || containsLower(s[1:], substr)))
}

func getCurrentPlatform() string {
	switch runtime.GOOS {
	case "darwin":
		return "macos"
	case "linux":
		return "linux"
	case "windows":
		return "windows"
	default:
		return "unknown"
	}
}


