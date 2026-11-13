package cmd

import (
	"euterpe/internal/ui"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const backupMenuTitle = "Available backups"

// TODO add option for argument
var restoreCommand = &cobra.Command{
	Use:   "restore",
	Short: "Restores your device's music library to a previous state from a backup folder",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var backupSelection string = ""
		directories, err := getBackupDirectories()
		if err != nil {
			return err
		}

		options, err := ui.ConstructSelectMenuOptions(directories)
		if err != nil {
			return err
		}

		backupMenu := ui.NewSelectMenu(backupMenuTitle, options, &backupSelection)
		backupMenu.Run()

		fmt.Printf("%s", backupSelection)

		return nil
	},
}

func getBackupDirectories() (map[string]string, error) {
	directoryMap := make(map[string]string)
	backupRoot := config.BackupRoot
	entries, err := os.ReadDir(backupRoot)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			directoryMap[entry.Name()] = filepath.Join(backupRoot, entry.Name())
		}
	}

	if len(directoryMap) == 0 {
		return nil, fmt.Errorf("no backups found\n")
	}

	return directoryMap, nil
}
