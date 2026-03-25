package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/juzhongsun/os-cleaner/internal/utils"
	"github.com/spf13/cobra"
)

var (
	activeScanPath string
)

var activeCmd = &cobra.Command{
	Use:   "active",
	Short: "Detect packages used by your projects",
	Long: `Detect which packages are actively used by your projects

Examples:
  os-cleaner active                           # Scan current directory
  os-cleaner active ~/projects/myapp         # Scan specific directory
  os-cleaner active --path ~/projects        # Scan multiple projects`,
	RunE: func(cmd *cobra.Command, args []string) error {
		searchPath := activeScanPath
		if searchPath == "" {
			wd, _ := os.Getwd()
			searchPath = wd
		}
		return detectActivePackages(searchPath)
	},
}

func init() {
	activeCmd.Flags().StringVarP(&activeScanPath, "path", "p", "", "Path to scan for projects")
	rootCmd.AddCommand(activeCmd)
}

type Project struct {
	Type     string
	Path     string
	Packages []string
}

func detectActivePackages(searchPath string) error {
	fmt.Println()
	fmt.Println(utils.Bold("═══════════════════════════════════════════════════════════════"))
	fmt.Println(utils.Bold("  Detecting Active Projects"))
	fmt.Println(utils.Bold("═══════════════════════════════════════════════════════════════"))
	fmt.Println()

	projects := findProjects(searchPath)

	if len(projects) == 0 {
		fmt.Println(utils.Yellow("No projects found in: " + searchPath))
		return nil
	}

	fmt.Printf(utils.Green("Found %d projects:\n\n"), len(projects))

	allPackages := make(map[string][]string)

	for _, p := range projects {
		fmt.Printf("  %s %s\n", utils.Bold(p.Type), utils.Dim(p.Path))

		for _, pkg := range p.Packages {
			if _, ok := allPackages[pkg]; !ok {
				allPackages[pkg] = []string{}
			}
			allPackages[pkg] = append(allPackages[pkg], p.Type)
		}
	}

	fmt.Println()
	fmt.Println(utils.Bold("═══════════════════════════════════════════════════════════════"))
	fmt.Println(utils.Bold("  Packages Used by Your Projects"))
	fmt.Println(utils.Bold("═══════════════════════════════════════════════════════════════"))
	fmt.Println()

	for pkg, types := range allPackages {
		fmt.Printf("  %s\n", utils.Bold(pkg))
		fmt.Printf("    %s\n", utils.Dim("Used in: "+join(types)))
	}

	fmt.Println()
	fmt.Println(utils.Bold("Tip: " + utils.Green("Compare this with 'os-cleaner inspect' to see what's not being used")))

	return nil
}

func findProjects(searchPath string) []Project {
	var projects []Project

	filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			// Skip common non-project directories
			skip := map[string]bool{
				"node_modules": true,
				".git":         true,
				"vendor":       true,
				"dist":         true,
				"build":        true,
				".cache":       true,
				".venv":        true,
				"venv":         true,
				"__pycache__":  true,
			}
			if skip[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// Detect project type by file
		dir := filepath.Dir(path)

		// Check if we already found a project in this directory
		for _, p := range projects {
			if p.Path == dir {
				return nil
			}
		}

		var p Project
		p.Path = dir

		switch info.Name() {
		case "package.json":
			p.Type = "Node.js"
			p.Packages = parsePackageJson(dir)
		case "go.mod":
			p.Type = "Go"
			p.Packages = parseGoMod(dir)
		case "requirements.txt":
			p.Type = "Python"
			p.Packages = parseRequirementsTxt(dir)
		case "Cargo.toml":
			p.Type = "Rust"
			p.Packages = parseCargoToml(dir)
		case "pom.xml":
			p.Type = "Maven"
			p.Packages = parsePomXml(dir)
		case "Gemfile":
			p.Type = "Ruby"
			p.Packages = parseGemfile(dir)
		case "mix.exs":
			p.Type = "Elixir"
			p.Packages = parseMixExs(dir)
		}

		if p.Type != "" {
			projects = append(projects, p)
		}

		return nil
	})

	return projects
}

func parsePackageJson(dir string) []string {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return nil
	}

	// Simple parse - just look for key:"value" patterns
	content := string(data)
	var packages []string

	// Check for dependencies
	for key := range map[string]string{} {
		_ = key
	}

	// Simple extraction using grep-like approach
	lines := splitLines(content)
	for _, line := range lines {
		line = trimSpace(line)
		if hasPrefix(line, `"dependencies"`) || hasPrefix(line, `"devDependencies"`) {
			continue
		}
		if contains(line, ":") && !contains(line, "//") {
			// Extract package name
			parts := splitOn(line, ":")
			if len(parts) >= 2 {
				name := trimQuotes(parts[0])
				if name != "" && !contains(name, "@") && len(name) > 2 {
					packages = append(packages, name)
				}
			}
		}
	}

	return packages
}

func parseGoMod(dir string) []string {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return nil
	}

	lines := splitLines(string(data))
	var packages []string

	for _, line := range lines {
		line = trimSpace(line)
		if hasPrefix(line, "require (") || hasPrefix(line, "require") {
			continue
		}
		if hasPrefix(line, "module ") || hasPrefix(line, "go ") {
			continue
		}
		if contains(line, " ") && !contains(line, "//") {
			parts := splitOn(line, " ")
			if len(parts) >= 2 {
				pkg := parts[0]
				if !contains(pkg, ".") && len(pkg) > 2 {
					packages = append(packages, pkg)
				}
			}
		}
	}

	return packages
}

func parseRequirementsTxt(dir string) []string {
	data, err := os.ReadFile(filepath.Join(dir, "requirements.txt"))
	if err != nil {
		return nil
	}

	lines := splitLines(string(data))
	var packages []string

	for _, line := range lines {
		line = trimSpace(line)
		if line == "" || hasPrefix(line, "#") || hasPrefix(line, "-") {
			continue
		}
		// Handle package==version or package>=version
		pkg := splitOn(line, "==")[0]
		pkg = splitOn(pkg, ">=")[0]
		pkg = splitOn(pkg, "<=")[0]
		pkg = splitOn(pkg, "~=")[0]
		if len(pkg) > 0 {
			packages = append(packages, pkg)
		}
	}

	return packages
}

func parseCargoToml(dir string) []string {
	data, err := os.ReadFile(filepath.Join(dir, "Cargo.toml"))
	if err != nil {
		return nil
	}

	content := string(data)
	var packages []string

	lines := splitLines(content)
	inDependencies := false

	for _, line := range lines {
		line = trimSpace(line)
		if hasPrefix(line, "[dependencies]") {
			inDependencies = true
			continue
		}
		if hasPrefix(line, "[") {
			inDependencies = false
			continue
		}
		if inDependencies && contains(line, "=") {
			parts := splitOn(line, "=")
			if len(parts) >= 2 {
				pkg := trimSpace(parts[0])
				if len(pkg) > 0 {
					packages = append(packages, pkg)
				}
			}
		}
	}

	return packages
}

func parsePomXml(dir string) []string {
	// Simplified - just check if pom.xml exists
	return []string{"Maven project"}
}

func parseGemfile(dir string) []string {
	data, err := os.ReadFile(filepath.Join(dir, "Gemfile"))
	if err != nil {
		return nil
	}

	lines := splitLines(string(data))
	var packages []string

	for _, line := range lines {
		line = trimSpace(line)
		if hasPrefix(line, "gem ") {
			parts := splitOn(line, " ")
			if len(parts) >= 2 {
				pkg := trimQuotes(parts[1])
				if len(pkg) > 0 {
					packages = append(packages, pkg)
				}
			}
		}
	}

	return packages
}

func parseMixExs(dir string) []string {
	data, err := os.ReadFile(filepath.Join(dir, "mix.exs"))
	if err != nil {
		return nil
	}

	content := string(data)
	var packages []string

	// Simple extraction
	lines := splitLines(content)
	for _, line := range lines {
		line = trimSpace(line)
		if hasPrefix(line, "{:") && contains(line, ",") {
			parts := splitOn(line, ",")
			if len(parts) >= 1 {
				pkg := trimQuotes(parts[0])
				pkg = trimPrefix(pkg, "{:")
				if len(pkg) > 0 {
					packages = append(packages, pkg)
				}
			}
		}
	}

	return packages
}

func join(items []string) string {
	result := ""
	for i, item := range items {
		if i > 0 {
			result += ", "
		}
		result += item
	}
	return result
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func splitOn(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[:len(substr)] == substr || contains(s[1:], substr)))
}

func trimQuotes(s string) string {
	s = trimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func trimPrefix(s, prefix string) string {
	if hasPrefix(s, prefix) {
		return s[len(prefix):]
	}
	return s
}
