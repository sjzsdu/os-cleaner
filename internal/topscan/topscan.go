package topscan

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/juzhongsun/os-cleaner/internal/utils"
)

type TopItem struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	IsDir     bool   `json:"is_dir"`
	FileCount int64  `json:"file_count,omitempty"`
}

type TopOptions struct {
	Path    string
	Top     int
	JSON    bool
	Verbose bool
}

const (
	WarnThreshold   = 100 * 1024 * 1024  // 100MB
	DangerThreshold = 1024 * 1024 * 1024 // 1GB
)

func Scan(opts TopOptions) error {
	targetPath := opts.Path
	if targetPath == "" {
		var err error
		targetPath, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	info, err := os.Stat(targetPath)
	if err != nil {
		return fmt.Errorf("cannot access %s: %w", targetPath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", targetPath)
	}

	entries, err := os.ReadDir(targetPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("Directory is empty")
		return nil
	}

	items := computeSizes(targetPath, entries, !opts.JSON)

	sort.Slice(items, func(i, j int) bool {
		return items[i].Size > items[j].Size
	})

	if opts.Top > 0 && len(items) > opts.Top {
		items = items[:opts.Top]
	}

	if opts.JSON {
		return outputJSON(items, targetPath)
	}

	outputTable(items, targetPath)
	return nil
}

func computeSizes(basePath string, entries []os.DirEntry, showProgress bool) []TopItem {
	items := make([]TopItem, len(entries))
	var wg sync.WaitGroup

	var progress *utils.Progress
	if showProgress {
		progress = utils.NewProgress(len(entries))
		progress.Start()
	}

	for i, entry := range entries {
		wg.Add(1)
		go func(idx int, e os.DirEntry) {
			defer wg.Done()

			fullPath := filepath.Join(basePath, e.Name())
			item := TopItem{
				Path: fullPath,
				Name: e.Name(),
			}

			info, err := e.Info()
			if err != nil {
				item.Size = 0
				items[idx] = item
				if progress != nil {
					progress.Increment(e.Name())
				}
				return
			}

			item.IsDir = info.IsDir()

			if info.IsDir() {
				size, count := getDirSizeAndCount(fullPath)
				item.Size = size
				item.FileCount = count
			} else {
				item.Size = info.Size()
			}

			items[idx] = item
		}(i, entry)
	}

	wg.Wait()
	return items
}

func getDirSizeAndCount(path string) (int64, int64) {
	var size atomic.Int64
	var count atomic.Int64

	var walk func(string)
	walk = func(dir string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}

		for _, entry := range entries {
			if entry.IsDir() {
				walk(filepath.Join(dir, entry.Name()))
			} else {
				info, err := entry.Info()
				if err != nil {
					continue
				}
				size.Add(info.Size())
				count.Add(1)
			}
		}
	}

	walk(path)
	return size.Load(), count.Load()
}

func outputTable(items []TopItem, basePath string) {
	var totalSize int64
	for _, item := range items {
		totalSize += item.Size
	}

	fmt.Println()
	fmt.Printf("%s %s\n", utils.Bold("Disk Usage for:"), basePath)
	fmt.Printf("%s %s\n", utils.Bold("Total:"), utils.Bold(utils.FormatSize(totalSize)))
	fmt.Println()

	fmt.Printf("  %-60s %10s %s\n", utils.Dim("Name"), utils.Dim("Size"), utils.Dim("Type"))
	fmt.Printf("  %s %s %s\n",
		utils.Dim("────────────────────────────────────────────────────────────"),
		utils.Dim("──────────"),
		utils.Dim("────────"))

	for i, item := range items {
		sizeStr := colorizeSize(item.Size)
		typeStr := "file"
		extra := ""

		if item.IsDir {
			typeStr = "dir"
			if item.FileCount > 0 {
				extra = fmt.Sprintf(" (%d files)", item.FileCount)
			}
		}

		indicator := ""
		if item.Size >= DangerThreshold {
			indicator = " ⚠️"
		} else if item.Size >= WarnThreshold {
			indicator = " ⚡"
		}

		fmt.Printf("  %2d. %-58s %10s %s%s%s\n",
			i+1,
			utils.TruncateString(item.Name, 58),
			sizeStr,
			utils.Dim(typeStr),
			utils.Dim(extra),
			indicator)
	}

	fmt.Println()

	var warnCount, dangerCount int
	for _, item := range items {
		if item.Size >= DangerThreshold {
			dangerCount++
		} else if item.Size >= WarnThreshold {
			warnCount++
		}
	}

	if dangerCount > 0 || warnCount > 0 {
		fmt.Printf("  %s\n", utils.Bold("Summary:"))
		if dangerCount > 0 {
			fmt.Printf("    %s %s\n", utils.Red(fmt.Sprintf("%d item(s) ≥ 1GB", dangerCount)), utils.Dim("(red)"))
		}
		if warnCount > 0 {
			fmt.Printf("    %s %s\n", utils.Yellow(fmt.Sprintf("%d item(s) ≥ 100MB", warnCount)), utils.Dim("(yellow)"))
		}
		fmt.Println()
	}
}

func colorizeSize(size int64) string {
	formatted := utils.FormatSize(size)
	if size >= DangerThreshold {
		return utils.Red(formatted)
	}
	if size >= WarnThreshold {
		return utils.Yellow(formatted)
	}
	return formatted
}

func outputJSON(items []TopItem, basePath string) error {
	output := struct {
		Path  string    `json:"path"`
		Items []TopItem `json:"items"`
	}{
		Path:  basePath,
		Items: items,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(data))
	return nil
}
