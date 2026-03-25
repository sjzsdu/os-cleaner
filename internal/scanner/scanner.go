package scanner

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/juzhongsun/os-cleaner/internal/registry"
	"github.com/juzhongsun/os-cleaner/internal/utils"
)

var spinChars = []string{"|", "/", "-", "\\"}
var spinIndex = 0

func printProgress(completed, total int, currentName string) {
	if completed == total {
		fmt.Println()
		return
	}

	fmt.Printf("\r  Scanning %s", spinChars[spinIndex%len(spinChars)])
}

// ScanResult represents the scan result for a category
type ScanResult struct {
	Category    string               `json:"category"`
	Name        string               `json:"name"`
	Path        string               `json:"path"`
	Size        int64                `json:"size"`
	SizeHuman   string               `json:"size_human"`
	FileCount   int64                `json:"file_count"`
	SafetyLevel registry.SafetyLevel `json:"safety_level"`
	LastAccess  time.Time            `json:"last_access,omitempty"`
	Error       string               `json:"error,omitempty"`
	Exists      bool                 `json:"exists"`
}

// ScanOptions defines options for scanning
type ScanOptions struct {
	Parallel  bool
	Verbose   bool
	JSON      bool
	Category  string
	ShowStale bool
	OlderThan time.Duration
}

// Scan performs the scan operation
func Scan(opts ScanOptions) error {
	var results []ScanResult
	var wg sync.WaitGroup

	categories := registry.GetCategoriesByPlatform()

	// Filter by category if specified
	if opts.Category != "" {
		filtered := []registry.CacheCategory{}
		for _, c := range categories {
			if c.ID == opts.Category {
				filtered = append(filtered, c)
				break
			}
		}
		categories = filtered
		if len(categories) == 0 {
			return fmt.Errorf("category not found: %s", opts.Category)
		}
	}

	// Create scan result channel
	resultsChan := make(chan ScanResult, len(categories))

	// Simple spinner - just print dots
	go func() {
		for {
			select {
			case <-time.After(500 * time.Millisecond):
				fmt.Print(".")
			}
		}
	}()

	// Wait for all scans to complete

	// Parallel scanning
	for _, cat := range categories {
		wg.Add(1)
		go func(cat registry.CacheCategory) {
			defer wg.Done()
			result := scanCategory(cat, opts)
			resultsChan <- result
		}(cat)
	}

	// Wait for all scans to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	for result := range resultsChan {
		results = append(results, result)
	}

	// Output results
	if opts.JSON {
		return outputJSON(results)
	}

	return outputTable(results, opts.Verbose, opts.ShowStale, opts.OlderThan)
}

func scanCategory(cat registry.CacheCategory, opts ScanOptions) ScanResult {
	result := ScanResult{
		Category:    cat.ID,
		Name:        cat.Name,
		SafetyLevel: cat.SafetyLevel,
	}

	var totalSize int64
	var totalFiles int64
	var exists bool

	// Scan each path in the category
	for _, pathRule := range cat.Paths {
		expandedPath := registry.ExpandPath(pathRule.Path)

		info, err := os.Stat(expandedPath)
		if err != nil {
			continue
		}

		exists = true

		if info.IsDir() {
			// Use du command for fast size calculation
			size, fileCount := fastGetSize(expandedPath)
			totalSize += size
			totalFiles += fileCount
		} else {
			totalSize += info.Size()
			totalFiles++
		}
	}

	result.Path = cat.Paths[0].Path
	result.Size = totalSize
	result.SizeHuman = utils.FormatSize(totalSize)
	result.FileCount = totalFiles
	result.Exists = exists

	// Only calculate time if needed (stale detection)
	if opts.ShowStale || opts.OlderThan > 0 {
		expandedPath := registry.ExpandPath(cat.Paths[0].Path)
		if info, err := os.Stat(expandedPath); err == nil && info.IsDir() {
			_, _, _, oldest := calculateDirSizeAndTime(expandedPath)
			if !oldest.IsZero() {
				result.LastAccess = oldest
			}
		}
	}

	return result
}

func fastGetSize(path string) (int64, int64) {
	// Use du -sk for fast size calculation (size in KB)
	// macOS compatible: -s summary, -k kilobytes
	cmd := exec.Command("du", "-sk", path)
	output, err := cmd.Output()
	if err != nil {
		return 0, 0
	}

	// du -sk output format: "12345\t/path/to/dir"
	parts := strings.Fields(string(output))
	if len(parts) >= 1 {
		sizeKB, _ := strconv.ParseInt(parts[0], 10, 64)
		// Convert KB to bytes
		size := sizeKB * 1024
		// Estimate file count based on size (rough approximation)
		// Average file size ~4KB for typical caches
		files := size / 4096
		if files < 1 {
			files = 1
		}
		return size, files
	}
	return 0, 0
}

func calculateDirSizeAndTime(path string) (int64, int64, time.Time, time.Time) {
	var size int64
	var files int64
	var latestAccess time.Time
	var oldestAccess time.Time

	filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !info.IsDir() {
			size += info.Size()
			files++
			if info.ModTime().After(latestAccess) {
				latestAccess = info.ModTime()
			}
			if oldestAccess.IsZero() || info.ModTime().Before(oldestAccess) {
				oldestAccess = info.ModTime()
			}
		}

		return nil
	})

	return size, files, latestAccess, oldestAccess
}

func outputJSON(results []ScanResult) error {
	output, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func outputTable(results []ScanResult, verbose, showStale bool, olderThan time.Duration) error {
	// Print header
	fmt.Println("\n" + utils.Bold("═══════════════════════════════════════════════════════════════"))

	title := "OS Cleaner Scan Results"
	if showStale {
		title = "Stale Caches (Not accessed recently)"
	} else if olderThan > 0 {
		title = fmt.Sprintf("Caches older than %v", olderThan)
	}
	fmt.Println(utils.Bold("                    " + title))
	fmt.Println(utils.Bold("═══════════════════════════════════════════════════════════════"))

	// Calculate threshold for stale
	var threshold time.Time
	if showStale {
		threshold = time.Now().AddDate(0, 0, -30) // 30 days
	} else if olderThan > 0 {
		threshold = time.Now().Add(-olderThan)
	}

	var totalSize int64
	var safeSize, cautionSize int64

	for _, r := range results {
		if !r.Exists {
			continue
		}

		// Filter by stale/olderThan
		if showStale || olderThan > 0 {
			if r.LastAccess.IsZero() || r.LastAccess.After(threshold) {
				continue
			}
		}

		totalSize += r.Size

		switch r.SafetyLevel {
		case registry.Safe:
			safeSize += r.Size
		case registry.Caution:
			cautionSize += r.Size
		}
	}

	fmt.Printf("\n")
	fmt.Printf("  %s %s\n", utils.Green("Total Cleanable:"), utils.Bold(utils.FormatSize(totalSize)))
	fmt.Printf("  %s %s\n", utils.Green("  - Safe:"), utils.FormatSize(safeSize))
	fmt.Printf("  %s %s\n", utils.Yellow("  - Caution:"), utils.FormatSize(cautionSize))

	if showStale {
		fmt.Printf("\n  %s\n", utils.Yellow("Showing caches not accessed in the last 30 days"))
	} else if olderThan > 0 {
		fmt.Printf("\n  %s\n", utils.Yellow(fmt.Sprintf("Showing caches not accessed in the last %v", olderThan)))
	}

	fmt.Printf("\n")

	// Print table header
	headerFormat := "  %-25s %12s %10s %8s %s\n"
	ageHeader := ""
	if showStale || olderThan > 0 {
		ageHeader = "Last Access"
	}
	fmt.Printf(headerFormat, "Category", "Size", "Files", "Level", ageHeader)
	fmt.Println("  " + utils.Dim("───────────────────────────────────────────────────────────────────────"))

	// Sort by size descending
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Size > results[i].Size {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// Print results
	for _, r := range results {
		if !r.Exists || r.Size == 0 {
			continue
		}

		// Filter by stale/olderThan
		if showStale || olderThan > 0 {
			if r.LastAccess.IsZero() || r.LastAccess.After(threshold) {
				continue
			}
		}

		levelStr := r.SafetyLevel.String()
		levelColor := utils.Green
		if r.SafetyLevel == registry.Caution {
			levelColor = utils.Yellow
		}

		ageStr := ""
		if showStale || olderThan > 0 {
			ageStr = r.LastAccess.Format("2006-01-02")
		}

		if ageStr != "" {
			fmt.Printf(headerFormat,
				utils.Bold(r.Name),
				utils.Bold(utils.FormatSize(r.Size)),
				utils.Dim(fmt.Sprintf("%d", r.FileCount)),
				levelColor(levelStr),
				utils.Dim(ageStr),
			)
		} else {
			fmt.Printf("  %-25s %12s %10s %8s\n",
				utils.Bold(r.Name),
				utils.Bold(utils.FormatSize(r.Size)),
				utils.Dim(fmt.Sprintf("%d", r.FileCount)),
				levelColor(levelStr),
			)
		}

		if verbose {
			fmt.Printf("    %s\n", utils.Dim(r.Path))
		}
	}

	fmt.Println("  " + utils.Dim("───────────────────────────────────────────────────────────────────────"))
	fmt.Printf("\n%s\n", utils.Green("Run 'os-cleaner clean <category>' to clean specific category"))
	fmt.Printf("%s\n", utils.Yellow("Run 'os-cleaner clean --safe' to clean all safe categories"))

	return nil
}
