package cmd

import (
	"strings"

	"github.com/juzhongsun/os-cleaner/internal/cleaner"
	"github.com/juzhongsun/os-cleaner/internal/registry"
	"github.com/spf13/cobra"
)

var (
	dryRun      bool
	safeMode    bool
	cautionMode bool
	categories  []string
	recoverable bool
)

var cleanCmd = &cobra.Command{
	Use:   "clean [category]",
	Short: "Clean specified cache category",
	Long: `Clean specified cache category or all categories

Examples:
  os-cleaner clean xcode                    # Clean Xcode caches
  os-cleaner clean docker                   # Clean Docker caches
  os-cleaner clean "macOS User Cache"       # Can use display name
  os-cleaner clean --safe                   # Clean all safe categories
  os-cleaner clean --dry-run                # Preview without deleting
  os-cleaner clean --recoverable            # Compress before deleting`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cleaner.CleanOptions{
			Categories:  categories,
			DryRun:      dryRun,
			SafeMode:    safeMode,
			CautionMode: cautionMode,
			Recoverable: recoverable,
			Verbose:     verbose,
			JSON:        jsonOutput,
		}

		if len(args) > 0 {
			// Try to find category by ID or name
			resolved := resolveCategory(args[0])
			if resolved == nil {
				// Show available categories and exit
				cmd.Println("Category not found:", args[0])
				cmd.Println("\nAvailable categories:")
				showCategories(cmd)
				return nil
			}
			opts.Categories = []string{resolved.ID}
		}

		return cleaner.Clean(opts)
	},
}

func init() {
	cleanCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without deleting")
	cleanCmd.Flags().BoolVar(&safeMode, "safe", false, "Only clean safe categories")
	cleanCmd.Flags().BoolVar(&cautionMode, "caution", false, "Clean safe + caution categories")
	cleanCmd.Flags().BoolVar(&recoverable, "recoverable", false, "Compress files before deletion for recovery")
	cleanCmd.Flags().StringSliceVarP(&categories, "category", "c", []string{}, "Specific categories to clean")
	rootCmd.AddCommand(cleanCmd)
}

func resolveCategory(input string) *registry.CacheCategory {
	input = strings.TrimSpace(input)
	inputLower := strings.ToLower(input)

	// First try exact ID match
	cat := registry.GetCategoryByID(input)
	if cat != nil {
		return cat
	}

	// Try exact name match (case insensitive)
	for _, c := range registry.GetCategoriesByPlatform() {
		if strings.ToLower(c.Name) == inputLower {
			return &c
		}
	}

	// Try partial match (ID or name contains input)
	for _, c := range registry.GetCategoriesByPlatform() {
		if strings.Contains(strings.ToLower(c.ID), inputLower) || strings.Contains(strings.ToLower(c.Name), inputLower) {
			return &c
		}
	}

	return nil
}

func showCategories(cmd *cobra.Command) {
	for _, c := range registry.GetCategoriesByPlatform() {
		cmd.Printf("  %-30s (%s)\n", c.ID, c.Name)
	}
}
