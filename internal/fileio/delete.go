package fileio

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

// TODO add progress bar
func DeleteFiles(filesToDelete []FileData, path string) error {
	cross := color.RedString("✕")

	for _, file := range filesToDelete {
		fileToDeletePath := filepath.Join(path, file.Name)
		err := os.Remove(fileToDeletePath)

		if err != nil {
			return fmt.Errorf("deleteFiles: %w", err)
		}

		fmt.Printf("%s deleted %s\n", cross, file.Name)
	}

	return nil
}
