package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/indium114/pricetag/internal"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var showAll bool
var gridMode bool

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func displayWidth(s string) int {
	plain := ansiRegex.ReplaceAllString(s, "")
	return len([]rune(plain))
}

type renderedEntry struct {
	display string
	width   int
}

var filelsCmd = &cobra.Command{
	Use:   "ls [directory]",
	Short: "List the contents of a directory, including tags",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) == 1 {
			dir = args[0]
		}

		db, _, err := internal.LoadDB()
		if err != nil {
			return err
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}

		var dirs []os.DirEntry
		var files []os.DirEntry

		for _, entry := range entries {
			name := entry.Name()

			if !showAll && strings.HasPrefix(name, ".") {
				continue
			}

			if entry.IsDir() {
				dirs = append(dirs, entry)
			} else {
				files = append(files, entry)
			}
		}

		sort.Slice(dirs, func(i, j int) bool {
			return dirs[i].Name() < dirs[j].Name()
		})
		sort.Slice(files, func(i, j int) bool {
			return files[i].Name() < files[j].Name()
		})

		var renderedEntries []renderedEntry

		// Directories
		for _, d := range dirs {
			name := d.Name()
			fullPath := filepath.Join(dir, name)

			absPath, err := internal.CanonicalPath(fullPath)
			if err != nil {
				continue
			}

			base := fmt.Sprintf(" %s", name)
			colorBase := internal.Colorize(base, "blue")

			var tagStrings []string
			if tags, ok := db.Paths[absPath]; ok {
				sort.Strings(tags)

				for _, tag := range tags {
					tagColor, exists := db.Tags[tag]
					if !exists {
						tagColor = "white"
					}

					tagText := fmt.Sprintf("[%s]", tag)
					tagStrings = append(tagStrings, internal.Colorize(tagText, internal.TagColor(tagColor)))
				}
			}

			var disp string
			if len(tagStrings) > 0 {
				disp = colorBase + " " + strings.Join(tagStrings, " ")
			} else {
				disp = colorBase
			}

			renderedEntries = append(renderedEntries, renderedEntry{display: disp, width: displayWidth(disp)})
		}

		// Files
		for _, f := range files {
			name := f.Name()
			fullPath := filepath.Join(dir, name)

			absPath, err := internal.CanonicalPath(fullPath)
			if err != nil {
				continue
			}

			ext := strings.TrimPrefix(filepath.Ext(name), ".")

			icon := ""
			fileColor := "white"

			if fileIcon, ok := db.Icons[ext]; ok {
				icon = fileIcon.Icon
				fileColor = string(fileIcon.Color)
			}

			// Base line, icon & filename
			base := fmt.Sprintf("%s %s", icon, name)
			colorBase := internal.Colorize(base, internal.TagColor(fileColor))

			// Tags
			var tagStrings []string
			if tags, ok := db.Paths[absPath]; ok {
				sort.Strings(tags)

				for _, tag := range tags {
					tagColor, exists := db.Tags[tag]
					// Shouldn't be possible, but just to be safe
					if !exists {
						tagColor = "white"
					}

					tagText := fmt.Sprintf("[%s]", tag)
					tagStrings = append(tagStrings, internal.Colorize(tagText, internal.TagColor(tagColor)))
				}
			}

			var disp string
			if len(tagStrings) > 0 {
				disp = colorBase + " " + strings.Join(tagStrings, "")
			} else {
				disp = colorBase
			}
			renderedEntries = append(renderedEntries, renderedEntry{display: disp, width: displayWidth(disp)})
		}

		if gridMode {
			maxWidth := 0
			for _, e := range renderedEntries {
				if e.width > maxWidth {
					maxWidth = e.width
				}
			}

			padding := 2
			cellWidth := maxWidth + padding

			termWidth := 80
			if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
				termWidth = w
			}

			cols := termWidth / cellWidth
			if cols < 1 {
				cols = 1
			}
			rows := (len(renderedEntries) + cols - 1) / cols

			for row := 0; row < rows; row++ {
				for col := 0; col < cols; col++ {
					idx := col*rows + row
					if idx >= len(renderedEntries) {
						continue
					}
					e := renderedEntries[idx]
					if col < cols-1 {
						fmt.Print(e.display + strings.Repeat(" ", cellWidth-e.width))
					} else {
						fmt.Print(e.display)
					}
				}
				fmt.Println()
			}
		} else {
			for _, e := range renderedEntries {
				fmt.Println(e.display)
			}
		}

		return nil
	},
}

func init() {
	fileCmd.AddCommand(filelsCmd)
	filelsCmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show hidden files")
	filelsCmd.Flags().BoolVar(&gridMode, "grid", false, "Display in grid layout")
}
