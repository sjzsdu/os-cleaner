package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	verbose    bool
	jsonOutput bool
	// Version is set at build time via ldflags (see .github/workflows/release.yml)
	Version = "dev"
)

var rootCmd = &cobra.Command{
	Use:   "os-cleaner",
	Short: "A cross-platform system cache cleaner",
	Long: `OS Cleaner - Clean system caches safely and efficiently

Supported platforms: macOS, Linux
Supported categories: System, Development Tools, Languages, Package Managers, Browsers

Examples:
  os-cleaner scan                    # Scan all cache categories
  os-cleaner scan --json            # Scan and output JSON
  os-cleaner list                    # List all cleanable categories
  os-cleaner clean xcode             # Clean Xcode caches
  os-cleaner clean --safe            # Clean all safe categories
  os-cleaner clean --dry-run         # Preview without deleting`,
	Version: Version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		startTime = time.Now()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		elapsed := time.Since(startTime)
		fmt.Println()
		fmt.Printf("  %s %s\n", Dim("Time elapsed:"), FormatDuration(elapsed))
	},
}

var startTime time.Time

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "JSON output")
}

func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Milliseconds()))
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}

func Dim(s string) string {
	return "\033[2m" + s + "\033[0m"
}
