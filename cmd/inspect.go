package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/juzhongsun/os-cleaner/internal/registry"
	"github.com/juzhongsun/os-cleaner/internal/utils"
	"github.com/spf13/cobra"
)

var (
	inspectAll bool
	inspectTop int
)

var inspectCmd = &cobra.Command{
	Use:   "inspect [category]",
	Short: "Inspect detailed contents of a cache category",
	Long: `Inspect what is stored in a specific cache category

Examples:
  os-cleaner inspect npm-cache           # Inspect npm cache
  os-cleaner inspect go-modules           # Inspect Go module cache
  os-cleaner inspect python-pyenv         # Inspect Python versions
  os-cleaner inspect --all                # Inspect all categories`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if inspectTop == 0 {
			inspectTop = 20
		}

		if len(args) > 0 {
			return inspectCategory(args[0], inspectTop)
		}

		if inspectAll {
			return inspectAllCategories(inspectTop)
		}

		return cmd.Help()
	},
}

func init() {
	inspectCmd.Flags().BoolVar(&inspectAll, "all", false, "Inspect all categories")
	inspectCmd.Flags().IntVarP(&inspectTop, "top", "n", 20, "Show top N items")
	rootCmd.AddCommand(inspectCmd)
}

func inspectCategory(categoryID string, top int) error {
	cat := registry.GetCategoryByID(categoryID)
	if cat == nil {
		return inspectByName(categoryID, top)
	}

	return doInspect(*cat, top)
}

func inspectByName(name string, top int) error {
	categories := registry.GetCategoriesByPlatform()

	for _, cat := range categories {
		if strings.Contains(cat.ID, name) || strings.Contains(cat.Name, name) {
			return doInspect(cat, top)
		}
	}

	// Try partial match
	var matched []registry.CacheCategory
	for _, cat := range categories {
		if strings.Contains(cat.ID, name) {
			matched = append(matched, cat)
		}
	}

	if len(matched) == 0 {
		return doInspectByPath(name, top)
	}

	if len(matched) == 1 {
		return doInspect(matched[0], top)
	}

	// Multiple matches - show list
	printMatchedCategories(matched)
	return nil
}

func printMatchedCategories(categories []registry.CacheCategory) {
	utils.PrintMatched("Multiple categories match:", categories)
}

func doInspectByPath(path string, top int) error {
	expandedPath := registry.ExpandPath(path)

	if !utils.PathExists(expandedPath) {
		// Try to find it under common locations
		home := utils.HomeDir()
		expandedPath = filepath.Join(home, path)
	}

	if !utils.PathExists(expandedPath) {
		expandedPath = filepath.Join(homeDirWithSlash(), strings.TrimPrefix(path, "~/"))
	}

	return inspectPath(expandedPath, path, top)
}

func homeDirWithSlash() string {
	home, _ := os.UserHomeDir()
	return home + "/"
}

func inspectAllCategories(top int) error {
	categories := registry.GetCategoriesByPlatform()

	utils.PrintSection("Inspecting All Categories", func() {
		for _, cat := range categories {
			doInspect(cat, top)
		}
	})

	return nil
}

func doInspect(cat registry.CacheCategory, top int) error {
	if len(cat.Paths) == 0 {
		return nil
	}

	pathRule := cat.Paths[0]
	expandedPath := registry.ExpandPath(pathRule.Path)

	if !utils.PathExists(expandedPath) {
		utils.PrintCategoryHeader(cat.Name, 0, utils.Dim("not found"))
		return nil
	}

	size, count := utils.GetDirSizeAndCount(expandedPath)
	utils.PrintCategoryHeader(cat.Name, size, formatCount(count))

	// Inspect based on category type
	switch {
	case isNPMCache(cat):
		inspectNPMCache(expandedPath, top)
	case isGoModules(cat):
		inspectGoModules(expandedPath, top)
	case isPythonPyenv(cat):
		inspectPythonPyenv(expandedPath, top)
	case isPythonPip(cat):
		inspectPythonPip(expandedPath, top)
	case isCargo(cat):
		inspectCargo(expandedPath, top)
	case isMaven(cat):
		inspectMaven(expandedPath, top)
	case isGradle(cat):
		inspectGradle(expandedPath, top)
	case isHomebrew(cat):
		inspectHomebrew(expandedPath, top)
	case isXcode(cat):
		inspectXcode(expandedPath, top)
	default:
		inspectGenericCache(expandedPath, top)
	}

	return nil
}

func inspectPath(path, name string, top int) error {
	if !utils.PathExists(path) {
		utils.PrintCategoryHeader(name, 0, utils.Red("not found"))
		return nil
	}

	size, count := utils.GetDirSizeAndCount(path)
	utils.PrintCategoryHeader(name, size, formatCount(count))
	inspectGenericCache(path, top)

	return nil
}

func inspectGenericCache(path string, top int) {
	cmd := exec.Command("du", "-h", path)
	output, _ := cmd.Output()

	// Show top directories
	cmd = exec.Command("du", "-h", "-d", "1", path)
	cmd.Dir = path
	output, _ = cmd.Output()

	lines := strings.Split(string(output), "\n")
	utils.PrintSubSection("Top subdirectories:")
	for i, line := range lines {
		if i >= top {
			break
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			utils.PrintItem(parts[1], parts[0])
		}
	}
}

func inspectNPMCache(path string, top int) {
	utils.PrintSubSection("NPM Cache contents (sample):")

	// Use npm cache ls to list packages
	cmd := exec.Command("npm", "cache", "ls")
	cmd.Env = append(cmd.Env, "HOME="+utils.HomeDir())
	output, err := cmd.Output()

	if err == nil {
		lines := strings.Split(string(output), "\n")
		count := 0
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			// Extract package name from URL
			parts := strings.Split(line, "/")
			if len(parts) > 0 {
				pkg := parts[len(parts)-1]
				if strings.Contains(pkg, ".tgz") {
					pkg = strings.TrimSuffix(pkg, ".tgz")
				}
				utils.PrintItem(pkg, "")
				count++
				if count >= top {
					break
				}
			}
		}
		if count == 0 {
			utils.PrintItem("Use `npm cache ls <package>` to inspect", "")
		}
	} else {
		utils.PrintItem("Use `npm cache verify` to verify", "")
	}

	// Show registry breakdown
	utils.ShowNpmRegistryBreakdown(path, top)
}

func inspectGoModules(path string, top int) {
	utils.PrintSubSection("Go Module cache by source:")

	cmd := exec.Command("du", "-h", "-d", "1", path)
	cmd.Dir = path
	output, _ := cmd.Output()

	lines := strings.Split(string(output), "\n")
	var results []struct {
		path string
		size string
	}

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			results = append(results, struct {
				path string
				size string
			}{parts[1], parts[0]})
		}
	}

	// Sort by size
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if parseSize(results[j].size) > parseSize(results[i].size) {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	for i, r := range results {
		if i >= top {
			break
		}
		utils.PrintItem(r.path, r.size)
	}
}

func inspectPythonPyenv(path string, top int) {
	utils.PrintSubSection("Installed Python versions:")

	cmd := exec.Command("ls", "-la", path)
	cmd.Dir = path
	output, _ := cmd.Output()

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "python") {
			parts := strings.Fields(line)
			if len(parts) >= 9 {
				utils.PrintItem(parts[8], "")
			}
		}
	}
}

func inspectPythonPip(path string, top int) {
	utils.PrintSubSection("pip cache (recent packages):")

	// Show wheel directory
	wheelPath := filepath.Join(path, "wheels")
	if utils.PathExists(wheelPath) {
		cmd := exec.Command("ls", "-lt", wheelPath)
		output, _ := cmd.Output()
		lines := strings.Split(string(output), "\n")
		count := 0
		for _, line := range lines {
			if count >= top {
				break
			}
			if strings.TrimSpace(line) == "" {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) >= 8 {
				utils.PrintItem(parts[7], "")
				count++
			}
		}
	}
}

func inspectCargo(path string, top int) {
	utils.PrintSubSection("Cargo registry by crate:")

	registryPath := filepath.Join(path, "registry", "cache")
	if !utils.PathExists(registryPath) {
		registryPath = filepath.Join(path, "registry")
	}

	if utils.PathExists(registryPath) {
		cmd := exec.Command("du", "-h", "-d", "2", registryPath)
		output, _ := cmd.Output()

		lines := strings.Split(string(output), "\n")
		count := 0
		for _, line := range lines {
			if count >= top {
				break
			}
			if strings.TrimSpace(line) == "" {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				subPath := parts[1]
				if strings.Contains(subPath, "github.com") || strings.Contains(subPath, "crates.io") {
					utils.PrintItem(subPath, parts[0])
					count++
				}
			}
		}
	}
}

func inspectMaven(path string, top int) {
	utils.PrintSubSection("Maven artifacts by group:")

	cmd := exec.Command("du", "-h", "-d", "1", path)
	cmd.Dir = path
	output, _ := cmd.Output()

	lines := strings.Split(string(output), "\n")
	count := 0
	for _, line := range lines {
		if count >= top {
			break
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			utils.PrintItem(parts[1], parts[0])
			count++
		}
	}
}

func inspectGradle(path string, top int) {
	utils.PrintSubSection("Gradle caches:")

	cmd := exec.Command("ls", "-la", path)
	output, _ := cmd.Output()

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "caches") || strings.Contains(line, "wrapper") {
			parts := strings.Fields(line)
			if len(parts) >= 8 {
				utils.PrintItem(parts[8], "")
			}
		}
	}
}

func inspectHomebrew(path string, top int) {
	utils.PrintSubSection("Homebrew cached bottles:")

	cmd := exec.Command("ls", "-lh", path)
	output, _ := cmd.Output()

	lines := strings.Split(string(output), "\n")
	count := 0
	for _, line := range lines {
		if count >= top {
			break
		}
		if strings.TrimSpace(line) == "" || strings.Contains(line, "total") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 8 {
			name := parts[8]
			if strings.HasSuffix(name, ".tar.gz") {
				name = strings.TrimSuffix(name, ".tar.gz")
				// Extract package name
				parts := strings.Split(name, "--")
				if len(parts) >= 2 {
					utils.PrintItem(parts[1], parts[0])
					count++
				}
			}
		}
	}
}

func inspectXcode(path string, top int) {
	utils.PrintSubSection("Xcode derived data:")

	cmd := exec.Command("ls", "-la", path)
	output, _ := cmd.Output()

	lines := strings.Split(string(output), "\n")
	count := 0
	for _, line := range lines {
		if count >= top {
			break
		}
		if strings.TrimSpace(line) == "" || strings.Contains(line, "total") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 8 {
			dirName := parts[8]
			if !strings.HasPrefix(dirName, ".") {
				// Get size of this derived data
				dirPath := filepath.Join(path, dirName)
				size, _ := utils.GetDirSize(dirPath)
				utils.PrintItem(dirName, utils.FormatSize(size))
				count++
			}
		}
	}
}

func isNPMCache(cat registry.CacheCategory) bool {
	return cat.ID == "npm-cache" || strings.Contains(cat.ID, "npm")
}

func isGoModules(cat registry.CacheCategory) bool {
	return cat.ID == "go-modules" || strings.Contains(cat.ID, "go-modules")
}

func isPythonPyenv(cat registry.CacheCategory) bool {
	return cat.ID == "python-pyenv"
}

func isPythonPip(cat registry.CacheCategory) bool {
	return cat.ID == "python-pip-cache"
}

func isCargo(cat registry.CacheCategory) bool {
	return cat.ID == "cargo-registry" || cat.ID == "cargo-git"
}

func isMaven(cat registry.CacheCategory) bool {
	return cat.ID == "maven-repository"
}

func isGradle(cat registry.CacheCategory) bool {
	return cat.ID == "gradle-cache"
}

func isHomebrew(cat registry.CacheCategory) bool {
	return cat.ID == "homebrew-cache"
}

func isXcode(cat registry.CacheCategory) bool {
	return strings.Contains(cat.ID, "xcode")
}

func formatCount(count int64) string {
	if count > 1000 {
		return utils.Dim(formatWithCommas(count))
	}
	return ""
}

func formatWithCommas(n int64) string {
	s := ""
	for i := 0; n > 0; i++ {
		if i > 0 && i%3 == 0 {
			s = "," + s
		}
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s + " items"
}

func parseSize(s string) int64 {
	s = strings.TrimSpace(s)
	s = strings.ToUpper(s)

	multiplier := int64(1)

	if strings.HasSuffix(s, "G") {
		multiplier = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "G")
	} else if strings.HasSuffix(s, "M") {
		multiplier = 1024 * 1024
		s = strings.TrimSuffix(s, "M")
	} else if strings.HasSuffix(s, "K") {
		multiplier = 1024
		s = strings.TrimSuffix(s, "K")
	}

	val := float64(0)
	for _, c := range s {
		if c >= '0' && c <= '9' || c == '.' {
			break
		}
	}

	_, err := fmt.Sscanf(s, "%f", &val)
	if err != nil {
		return 0
	}

	return int64(float64(multiplier) * val)
}
