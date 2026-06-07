package cmd

import (
	"os"

	"github.com/indium114/pricetag/internal"
	"github.com/indium114/slag"
	"github.com/spf13/cobra"
)

var filecleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Purge non-existent files from database",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, path, err := internal.LoadDB()
		if err != nil {
			return err
		}

		var removed []string
		for filePath := range db.Paths {
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				delete(db.Paths, filePath)
				removed = append(removed, filePath)
				slag.Clean("%s\n", filePath)
			}
		}

		if err := internal.SaveDB(db, path); err != nil {
			return err
		}

		if len(removed) == 0 {
			slag.Ok("No stale database entries found")
			return nil
		}

		slag.Ok("Removed %d stale entr%s\n", len(removed), map[bool]string{true: "ies", false: "y"}[len(removed) != 1])
		return nil
	},
}

func init() {
	fileCmd.AddCommand(filecleanCmd)
}
