package formatter

import (
	"encoding/json"
	"fmt"

	"github.com/juzhongsun/os-cleaner/internal/registry"
	"github.com/juzhongsun/os-cleaner/internal/utils"
)

// ListCategories lists all available cache categories
func ListCategories(categories []registry.CacheCategory, jsonOutput bool) error {
	if jsonOutput {
		return listCategoriesJSON(categories)
	}
	return listCategoriesTable(categories)
}

func listCategoriesJSON(categories []registry.CacheCategory) error {
	output, err := json.MarshalIndent(categories, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func listCategoriesTable(categories []registry.CacheCategory) error {
	fmt.Println("\n" + utils.Bold("═══════════════════════════════════════════════════════════════"))
	fmt.Println(utils.Bold("                   Available Cache Categories                  "))
	fmt.Println(utils.Bold("═══════════════════════════════════════════════════════════════\n"))

	fmt.Printf("  %-25s %-12s %-10s %s\n",
		utils.Bold("ID"),
		utils.Bold("Safety"),
		utils.Bold("Platform"),
		utils.Bold("Description"))
	fmt.Println(utils.Dim("  " + "───────────────────────────────────────────────────────────────────"))

	for _, cat := range categories {
		levelColor := utils.Green
		if cat.SafetyLevel == registry.Caution {
			levelColor = utils.Yellow
		} else if cat.SafetyLevel == registry.Dangerous {
			levelColor = utils.Red
		}

		platforms := ""
		if len(cat.Platforms) > 0 {
			platforms = cat.Platforms[0]
			if len(cat.Platforms) > 1 {
				platforms = "multi"
			}
		}

		fmt.Printf("  %-25s %s %-10s %s\n",
			utils.Bold(cat.ID),
			levelColor(cat.SafetyLevel.String()),
			utils.Dim(platforms),
			utils.Dim(utils.TruncateString(cat.Description, 40)),
		)
	}

	fmt.Println(utils.Dim("  " + "───────────────────────────────────────────────────────────────────"))
	fmt.Println()

	return nil
}
