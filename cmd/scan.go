package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/juzhongsun/os-cleaner/internal/scanner"
	"github.com/spf13/cobra"
)

var (
	scanStale     bool
	scanOlderThan string
	scanMinSize   string
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan all cache categories",
	Long:  "Scan all cache categories and display their sizes",
	RunE: func(cmd *cobra.Command, args []string) error {
		minSize, err := parseSize(scanMinSize)
		if err != nil {
			return fmt.Errorf("invalid --min-size: %w", err)
		}
		opts := scanner.ScanOptions{
			Verbose:   verbose,
			JSON:      jsonOutput,
			ShowStale: scanStale,
			OlderThan: parseDuration(scanOlderThan),
			MinSize:   minSize,
		}
		return scanner.Scan(opts)
	},
}

func init() {
	scanCmd.Flags().BoolVar(&scanStale, "stale", false, "Show caches not accessed in the last 30 days")
	scanCmd.Flags().StringVar(&scanOlderThan, "older-than", "", "Show caches older than (e.g., 30d, 90d, 6m)")
	scanCmd.Flags().StringVar(&scanMinSize, "min-size", "", "Minimum cache size to show (e.g., 10MB, 1GB)")
	rootCmd.AddCommand(scanCmd)
}

// parseSize parses a human-readable size string like "10MB", "1GB", "500KB" into bytes.
func parseSize(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}
	s = strings.ToUpper(strings.TrimSpace(s))
	var multiplier int64 = 1
	switch {
	case strings.HasSuffix(s, "TB"):
		multiplier = 1024 * 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "TB")
	case strings.HasSuffix(s, "GB"):
		multiplier = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "GB")
	case strings.HasSuffix(s, "MB"):
		multiplier = 1024 * 1024
		s = strings.TrimSuffix(s, "MB")
	case strings.HasSuffix(s, "KB"):
		multiplier = 1024
		s = strings.TrimSuffix(s, "KB")
	case strings.HasSuffix(s, "B"):
		s = strings.TrimSuffix(s, "B")
	}
	val, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("unable to parse size %q", s)
	}
	return val * multiplier, nil
}

func parseDuration(s string) time.Duration {
	if s == "" {
		return 0
	}

	var days int
	_, err := fmt.Sscanf(s, "%dd", &days)
	if err != nil {
		_, err = fmt.Sscanf(s, "%d", &days)
	}
	if days > 0 {
		return time.Duration(days) * 24 * time.Hour
	}
	return 0
}
