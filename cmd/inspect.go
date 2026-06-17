package cmd

import (
	"github.com/juzhongsun/os-cleaner/internal/inspect"
	"github.com/spf13/cobra"
)

var (
	inspectAll bool
	inspectTop int
)

var inspectCmd = &cobra.Command{
	Use:   "inspect [category]",
	Short: "Inspect detailed contents of a cache category",
	Long: `Inspect what is stored in a specific cache category

Examples:
  os-cleaner inspect npm-cache           # Inspect npm cache
  os-cleaner inspect go-modules           # Inspect Go module cache
  os-cleaner inspect python-pyenv         # Inspect Python versions
  os-cleaner inspect --all                # Inspect all categories`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := inspect.InspectOptions{
			Top: inspectTop,
			All: inspectAll,
		}
		if len(args) > 0 {
			opts.Category = args[0]
		}
		return inspect.Run(opts)
	},
}

func init() {
	inspectCmd.Flags().BoolVar(&inspectAll, "all", false, "Inspect all categories")
	inspectCmd.Flags().IntVarP(&inspectTop, "top", "n", 20, "Show top N items")
	rootCmd.AddCommand(inspectCmd)
}
