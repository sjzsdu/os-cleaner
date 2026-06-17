package inspect

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/juzhongsun/os-cleaner/internal/registry"
	"github.com/juzhongsun/os-cleaner/internal/utils"
)

const defaultTopN = 20

// InspectOptions defines options for inspection
type InspectOptions struct {
	Category string
	Top      int
	All      bool
}

// Run performs the inspect operation based on options
func Run(opts InspectOptions) error {
	top := opts.Top
	if top == 0 {
		top = defaultTopN
	}

	if opts.Category != "" {
		return inspectCategory(opts.Category, top)
	}

	if opts.All {
		return inspectAllCategories(top)
	}

	return fmt.Errorf("specify a category or use --all")
}

func inspectCategory(categoryID string, top int) error {
	cat := registry.GetCategoryByID(categoryID)
	if cat == nil {
		return inspectByName(categoryID, top)
	}

	return doInspect(*cat, top)
}

func inspectByName(name string, top int) error {
	matched := registry.SearchCategories(name)
	if len(matched) == 0 {
		return doInspectByPath(name, top)
	}
	if len(matched) == 1 {
		return doInspect(matched[0], top)
	}
	printMatchedCategories(matched)
	return nil
}

func printMatchedCategories(categories []registry.CacheCategory) {
	printMatched("Multiple categories match:", categories)
}

func doInspectByPath(path string, top int) error {
	expandedPath := utils.ExpandPath(path)

	if !utils.PathExists(expandedPath) {
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

	printSection("Inspecting All Categories", func() {
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
	expandedPath := utils.ExpandPath(pathRule.Path)

	if !utils.PathExists(expandedPath) {
		printCategoryHeader(cat.Name, 0, utils.Dim("not found"))
		return nil
	}

	size, count := getDirSizeAndCount(expandedPath)
	printCategoryHeader(cat.Name, size, formatCount(count))

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
		printCategoryHeader(name, 0, utils.Red("not found"))
		return nil
	}

	size, count := getDirSizeAndCount(path)
	printCategoryHeader(name, size, formatCount(count))
	inspectGenericCache(path, top)

	return nil
}

func inspectGenericCache(path string, top int) {
	cmd := exec.Command("du", "-h", "-d", "1", path)
	cmd.Dir = path
	output, _ := cmd.Output()

	lines := strings.Split(string(output), "\n")
	printSubSection("Top subdirectories:")
	for i, line := range lines {
		if i >= top {
			break
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			printItem(parts[1], parts[0])
		}
	}
}

func inspectNPMCache(path string, top int) {
	printSubSection("NPM Cache contents (sample):")

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
			parts := strings.Split(line, "/")
			if len(parts) > 0 {
				pkg := parts[len(parts)-1]
				if strings.Contains(pkg, ".tgz") {
					pkg = strings.TrimSuffix(pkg, ".tgz")
				}
				printItem(pkg, "")
				count++
				if count >= top {
					break
				}
			}
		}
		if count == 0 {
			printItem("Use `npm cache ls <package>` to inspect", "")
		}
	} else {
		printItem("Use `npm cache verify` to verify", "")
	}

	showNpmRegistryBreakdown(path, top)
}

func inspectGoModules(path string, top int) {
	printSubSection("Go Module cache by source:")

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
		printItem(r.path, r.size)
	}
}

func inspectPythonPyenv(path string, top int) {
	_ = top
	printSubSection("Installed Python versions:")

	cmd := exec.Command("ls", "-la", path)
	cmd.Dir = path
	output, _ := cmd.Output()

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "python") {
			parts := strings.Fields(line)
			if len(parts) >= 9 {
				printItem(parts[8], "")
			}
		}
	}
}

func inspectPythonPip(path string, top int) {
	printSubSection("pip cache (recent packages):")

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
				printItem(parts[7], "")
				count++
			}
		}
	}
}

func inspectCargo(path string, top int) {
	printSubSection("Cargo registry by crate:")

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
					printItem(subPath, parts[0])
					count++
				}
			}
		}
	}
}

func inspectMaven(path string, top int) {
	printSubSection("Maven artifacts by group:")

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
			printItem(parts[1], parts[0])
			count++
		}
	}
}

func inspectGradle(path string, top int) {
	_ = top
	printSubSection("Gradle caches:")

	cmd := exec.Command("ls", "-la", path)
	output, _ := cmd.Output()

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "caches") || strings.Contains(line, "wrapper") {
			parts := strings.Fields(line)
			if len(parts) >= 8 {
				printItem(parts[8], "")
			}
		}
	}
}

func inspectHomebrew(path string, top int) {
	printSubSection("Homebrew cached bottles:")

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
				parts := strings.Split(name, "--")
				if len(parts) >= 2 {
					printItem(parts[1], parts[0])
					count++
				}
			}
		}
	}
}

func inspectXcode(path string, top int) {
	printSubSection("Xcode derived data:")

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
				dirPath := filepath.Join(path, dirName)
				size, _ := utils.GetDirSize(dirPath)
				printItem(dirName, utils.FormatSize(size))
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

func getDirSizeAndCount(path string) (int64, int64) {
	var size int64
	var count int64

	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
			count++
		}
		return nil
	})

	return size, count
}
