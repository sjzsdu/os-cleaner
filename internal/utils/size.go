package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FormatSize formats bytes into human readable string
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Bold returns bold text
func Bold(s string) string {
	return "\033[1m" + s + "\033[0m"
}

// Green returns green colored text
func Green(s string) string {
	return "\033[32m" + s + "\033[0m"
}

// Yellow returns yellow colored text
func Yellow(s string) string {
	return "\033[33m" + s + "\033[0m"
}

// Red returns red colored text
func Red(s string) string {
	return "\033[31m" + s + "\033[0m"
}

// Dim returns dimmed text
func Dim(s string) string {
	return "\033[2m" + s + "\033[0m"
}

// HomeDir returns the user's home directory
func HomeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

// ExpandPath expands ~ and environment variables
func ExpandPath(path string) string {
	if path == "" {
		return ""
	}

	// Handle ~ expansion
	if path == "~" {
		return HomeDir()
	}

	if len(path) >= 2 && path[:2] == "~/" {
		return HomeDir() + path[1:]
	}

	return os.ExpandEnv(path)
}

// PathExists checks if a path exists
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir checks if path is a directory
func IsDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// GetDirSize calculates total size of a directory
func GetDirSize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// CompressDir compresses a directory into a tar.gz archive
func CompressDir(srcPath, destPath string) error {
	cmd := exec.Command("tar", "-czf", destPath, "-C", filepath.Dir(srcPath), filepath.Base(srcPath))
	return cmd.Run()
}

// RemovePath removes a file or directory
func RemovePath(path string) error {
	return os.RemoveAll(path)
}

// TruncateString truncates a string to maxLength
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength-3] + "..."
}

// SanitizeCategoryName sanitizes a category name for use in filenames
func SanitizeCategoryName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "/", "-")
	return name
}
