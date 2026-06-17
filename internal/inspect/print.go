package inspect

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/juzhongsun/os-cleaner/internal/registry"
	"github.com/juzhongsun/os-cleaner/internal/utils"
)

func printMatched(msg string, categories []registry.CacheCategory) {
	fmt.Println(utils.Bold(msg))
	for _, c := range categories {
		fmt.Printf("  - %s (%s)\n", c.ID, c.Name)
	}
}

func printSection(name string, fn func()) {
	fmt.Println()
	fmt.Println(utils.Bold("═══════════════════════════════════════════════════════════════"))
	fmt.Println(utils.Bold("  " + name))
	fmt.Println(utils.Bold("═══════════════════════════════════════════════════════════════"))
	fmt.Println()
	fn()
}

func printCategoryHeader(name string, size int64, extra string) {
	fmt.Println()
	fmt.Printf("%s %s", utils.Bold(name), utils.Bold(utils.FormatSize(size)))
	if extra != "" {
		fmt.Printf(" %s", extra)
	}
	fmt.Println()
}

func printSubSection(name string) {
	fmt.Println()
	fmt.Println(utils.Dim("  " + name))
}

func printItem(name, size string) {
	if size != "" {
		fmt.Printf("    %-50s %s\n", name, utils.Dim(size))
	} else {
		fmt.Printf("    %s\n", name)
	}
}

func showNpmRegistryBreakdown(path string, top int) {
	registryPath := filepath.Join(path, "_cacache", "content-v2", "sha512")
	if !utils.PathExists(registryPath) {
		return
	}

	cmd := exec.Command("du", "-h", "-d", "1", registryPath)
	output, _ := cmd.Output()

	lines := strings.Split(string(output), "\n")
	printSubSection("By registry source:")
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
