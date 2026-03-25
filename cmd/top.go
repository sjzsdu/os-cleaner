package cmd

import (
	"github.com/juzhongsun/os-cleaner/internal/topscan"
	"github.com/spf13/cobra"
)

var topCount int

var topCmd = &cobra.Command{
	Use:   "top [path]",
	Short: "Show largest files and directories",
	Long:  "Scan a directory and display the largest files and directories by size. Highlights items ≥100MB in yellow and ≥1GB in red.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetPath := ""
		if len(args) > 0 {
			targetPath = args[0]
		}

		opts := topscan.TopOptions{
			Path:    targetPath,
			Top:     topCount,
			JSON:    jsonOutput,
			Verbose: verbose,
		}
		return topscan.Scan(opts)
	},
}

func init() {
	topCmd.Flags().IntVarP(&topCount, "top", "n", 20, "Number of items to show")
	rootCmd.AddCommand(topCmd)
}
