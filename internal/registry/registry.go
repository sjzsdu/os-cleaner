package registry

import (
	"os"
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
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Platforms   []string    `json:"platforms"` // "macos", "linux", "all"
	SafetyLevel SafetyLevel `json:"safety_level"`
	Paths       []PathRule  `json:"paths"`
	CleanCmd    string      `json:"clean_cmd,omitempty"` // Optional cleanup command
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

// ExpandPath expands ~ and environment variables in path
func ExpandPath(path string) string {
	if path == "" {
		return ""
	}

	// Handle ~ expansion
	if path == "~" {
		home, _ := os.UserHomeDir()
		return home
	}

	if len(path) >= 2 && path[:2] == "~/" {
		home, _ := os.UserHomeDir()
		return home + path[1:]
	}

	// Handle $HOME
	home := os.Getenv("HOME")
	if home != "" && len(path) >= len(home)+2 && path[:len(home)] == home {
		return os.ExpandEnv(path)
	}

	return os.ExpandEnv(path)
}
