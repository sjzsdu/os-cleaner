package cmd

import (
	"os"

	"github.com/juzhongsun/os-cleaner/internal/active"
	"github.com/spf13/cobra"
)

var (
	activeScanPath string
)

var activeCmd = &cobra.Command{
	Use:   "active",
	Short: "Detect packages used by your projects",
	Long: `Detect which packages are actively used by your projects

Examples:
  os-cleaner active                           # Scan current directory
  os-cleaner active ~/projects/myapp         # Scan specific directory
  os-cleaner active --path ~/projects        # Scan multiple projects`,
	RunE: func(cmd *cobra.Command, args []string) error {
		searchPath := activeScanPath
		if searchPath == "" {
			wd, _ := os.Getwd()
			searchPath = wd
		}
		return active.Run(searchPath)
	},
}

func init() {
	activeCmd.Flags().StringVarP(&activeScanPath, "path", "p", "", "Path to scan for projects")
	rootCmd.AddCommand(activeCmd)
}
