package cmd

import (
	"github.com/juzhongsun/os-cleaner/internal/formatter"
	"github.com/juzhongsun/os-cleaner/internal/registry"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all cleanable categories",
	Long:  "List all cleanable cache categories with their paths",
	RunE: func(cmd *cobra.Command, args []string) error {
		categories := registry.GetAllCategories()
		return formatter.ListCategories(categories, jsonOutput)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
