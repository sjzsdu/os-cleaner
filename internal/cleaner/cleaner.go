package cleaner

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/juzhongsun/os-cleaner/internal/registry"
	"github.com/juzhongsun/os-cleaner/internal/utils"
)

// CleanOptions defines options for cleaning
type CleanOptions struct {
	Categories  []string
	DryRun      bool
	SafeMode    bool
	CautionMode bool
	Recoverable bool
	Verbose     bool
	JSON        bool
}

// CleanResult represents the result of a clean operation
type CleanResult struct {
	Category    string `json:"category"`
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	SizeHuman   string `json:"size_human"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
	Recoverable string `json:"recoverable,omitempty"` // Path to compressed archive if recoverable
}

// Clean performs the clean operation
func Clean(opts CleanOptions) error {
	// Determine which categories to clean
	var categoriesToClean []registry.CacheCategory

	if opts.SafeMode {
		categoriesToClean = registry.GetSafeCategories()
	} else if opts.CautionMode {
		categoriesToClean = registry.GetCautionCategories()
	} else if len(opts.Categories) > 0 {
		for _, catID := range opts.Categories {
			cat := registry.GetCategoryByID(catID)
			if cat != nil {
				categoriesToClean = append(categoriesToClean, *cat)
			} else {
				fmt.Printf("Warning: category not found: %s\n", catID)
			}
		}
	} else {
		// Default: use safe categories
		categoriesToClean = registry.GetSafeCategories()
	}

	if len(categoriesToClean) == 0 {
		return fmt.Errorf("no categories to clean")
	}

	// Perform cleaning
	results := cleanCategories(categoriesToClean, opts)

	// Output results
	if opts.JSON {
		return outputJSON(results)
	}

	return outputTable(results, opts.DryRun)
}

func cleanCategories(categories []registry.CacheCategory, opts CleanOptions) []CleanResult {
	var results []CleanResult

	for _, cat := range categories {
		result := cleanCategory(cat, opts)
		results = append(results, result)
	}

	return results
}

func cleanCategory(cat registry.CacheCategory, opts CleanOptions) CleanResult {
	result := CleanResult{
		Category: cat.ID,
		Status:   "skipped",
	}

	var totalSize int64
	var deletedSize int64
	var permissionErrors []string

	for _, pathRule := range cat.Paths {
		expandedPath := registry.ExpandPath(pathRule.Path)

		if !utils.PathExists(expandedPath) {
			continue
		}

		result.Path = pathRule.Path

		// Calculate size
		if utils.IsDir(expandedPath) {
			size, _ := utils.GetDirSize(expandedPath)
			totalSize += size
		} else {
			info, _ := os.Stat(expandedPath)
			totalSize += info.Size()
		}

		// Dry run - don't actually delete
		if opts.DryRun {
			result.Status = "dry-run"
			continue
		}

		// Recoverable mode - compress before deleting
		if opts.Recoverable {
			archivePath := createRecoveryArchive(expandedPath, cat.ID)
			if archivePath != "" {
				result.Recoverable = archivePath
			}
		}

		delSize, errs := safeRemoveWithDetails(expandedPath)
		deletedSize += delSize
		permissionErrors = append(permissionErrors, errs...)
	}

	result.Size = totalSize
	result.SizeHuman = utils.FormatSize(totalSize)

	if opts.DryRun {
		result.Status = "dry-run"
	} else if len(permissionErrors) > 0 {
		if deletedSize > 0 {
			result.Status = "partial"
			if len(permissionErrors) <= 3 {
				result.Error = fmt.Sprintf("Some files not deleted due to permissions: %s", strings.Join(permissionErrors, "; "))
			} else {
				result.Error = fmt.Sprintf("%d files not deleted due to permissions. Use sudo for full cleanup: sudo os-cleaner clean %s", len(permissionErrors), cat.ID)
			}
		} else {
			result.Status = "error"
			result.Error = fmt.Sprintf("All files require sudo/root permission. Try: sudo os-cleaner clean %s", cat.ID)
		}
	} else if totalSize > 0 {
		result.Status = "cleaned"
	} else {
		result.Status = "skipped"
	}

	return result
}

func createRecoveryArchive(sourcePath, category string) string {
	// Create recovery directory
	recoverDir := filepath.Join(utils.HomeDir(), ".os-cleaner", "recover")
	if err := os.MkdirAll(recoverDir, 0755); err != nil {
		fmt.Printf("Warning: failed to create recovery directory: %v\n", err)
		return ""
	}

	// Generate archive name
	timestamp := time.Now().Format("2006-01-02_150405")
	archiveName := fmt.Sprintf("%s_%s.tar.gz", category, timestamp)
	archivePath := filepath.Join(recoverDir, archiveName)

	// Compress the directory
	if err := utils.CompressDir(sourcePath, archivePath); err != nil {
		fmt.Printf("Warning: failed to compress %s: %v\n", sourcePath, err)
		return ""
	}

	fmt.Printf("  Created recovery archive: %s\n", archivePath)
	return archivePath
}

func outputJSON(results []CleanResult) error {
	output, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func outputTable(results []CleanResult, dryRun bool) error {
	if dryRun {
		fmt.Println("\n" + utils.Yellow("═══════════════════════════════════════════════════════════════"))
		fmt.Println(utils.Yellow("                      DRY RUN MODE                              "))
		fmt.Println(utils.Yellow("═══════════════════════════════════════════════════════════════"))
		fmt.Println(utils.Yellow("  No files will be deleted. This is a preview.\n"))
	}

	fmt.Println("\n" + utils.Bold("═══════════════════════════════════════════════════════════════"))
	fmt.Println(utils.Bold("                      Clean Results                         "))
	fmt.Println(utils.Bold("═══════════════════════════════════════════════════════════════\n"))

	var totalCleaned int64

	for _, r := range results {
		statusColor := utils.Green
		if r.Status == "error" {
			statusColor = utils.Red
		} else if r.Status == "dry-run" {
			statusColor = utils.Yellow
		} else if r.Status == "partial" {
			statusColor = utils.Yellow
		} else if r.Status == "skipped" {
			statusColor = utils.Dim
		}

		statusText := r.Status
		if r.Status == "skipped" && r.Size == 0 {
			statusText = "not found"
		}

		fmt.Printf("  %s\n", utils.Bold(r.Category))
		fmt.Printf("    Path: %s\n", utils.Dim(r.Path))
		fmt.Printf("    Size: %s\n", utils.Bold(r.SizeHuman))
		fmt.Printf("    Status: %s\n", statusColor(statusText))

		if r.Recoverable != "" {
			fmt.Printf("    Recovery: %s\n", utils.Green(r.Recoverable))
		}

		if r.Error != "" {
			fmt.Printf("    Error: %s\n", utils.Red(r.Error))
		}

		fmt.Println()

		if r.Status == "cleaned" || r.Status == "dry-run" || r.Status == "partial" {
			totalCleaned += r.Size
		}
	}

	fmt.Println(utils.Dim("───────────────────────────────────────────────────────────────"))
	fmt.Printf("\n  %s: %s\n", utils.Bold("Total"), utils.Bold(utils.FormatSize(totalCleaned)))

	if dryRun {
		fmt.Printf("\n%s\n", utils.Yellow("Run without --dry-run to actually delete files"))
	}

	if !dryRun {
		fmt.Printf("\n%s\n", utils.Green("Cleanup complete!"))
		if totalCleaned > 0 {
			fmt.Printf("%s: %s\n", utils.Green("Freed"), utils.Bold(utils.FormatSize(totalCleaned)))
		}
	}

	return nil
}

func safeRemoveWithDetails(path string) (int64, []string) {
	var deletedSize int64
	var errors []string

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, []string{err.Error()}
	}

	if info.IsDir() {
		size, _ := utils.GetDirSize(path)
		deletedSize = size

		err := os.RemoveAll(path)
		if err != nil {
			if strings.Contains(err.Error(), "permission") || strings.Contains(err.Error(), "Operation not permitted") {
				errors = append(errors, path)
			} else if !strings.Contains(err.Error(), "no such file") {
				errors = append(errors, err.Error())
			}
		}
	} else {
		deletedSize = info.Size()
		err := os.Remove(path)
		if err != nil {
			if strings.Contains(err.Error(), "permission") {
				errors = append(errors, path)
			} else if !strings.Contains(err.Error(), "no such file") {
				errors = append(errors, err.Error())
			}
		}
	}

	return deletedSize, errors
}

func safeRemove(path string) error {
	cmd := exec.Command("rm", "-rf", path)
	err := cmd.Run()
	return err
}
