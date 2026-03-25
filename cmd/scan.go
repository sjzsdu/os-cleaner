package cmd

import (
	"fmt"
	"time"

	"github.com/juzhongsun/os-cleaner/internal/scanner"
	"github.com/spf13/cobra"
)

var (
	scanStale     bool
	scanOlderThan string
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan all cache categories",
	Long:  "Scan all cache categories and display their sizes",
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := scanner.ScanOptions{
			Parallel:  true,
			Verbose:   verbose,
			JSON:      jsonOutput,
			ShowStale: scanStale,
			OlderThan: parseDuration(scanOlderThan),
		}
		return scanner.Scan(opts)
	},
}

func init() {
	scanCmd.Flags().BoolVar(&scanStale, "stale", false, "Show caches not accessed in the last 30 days")
	scanCmd.Flags().StringVar(&scanOlderThan, "older-than", "", "Show caches older than (e.g., 30d, 90d, 6m)")
	rootCmd.AddCommand(scanCmd)
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
