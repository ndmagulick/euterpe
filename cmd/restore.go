package cmd

import (
	"euterpe/internal/backup"
	"euterpe/ui"
	"fmt"

	"github.com/spf13/cobra"
)

const backupMenuTitle = "Available backups"
const restoreConfirmationTitle = "Are you sure you want to restore?"
const restoreConfirmationDescription = "Restoring from a backup will wipe all audio files currently on your device, and copy all files in the backup folder.\nThis operation will also purge the cache which will slow down your next sync."

// TODO add option for argument
var restoreCommand = &cobra.Command{
	Use:   "restore",
	Short: "Restores your device's music library to a previous state from a backup folder",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		var confirm bool
		confirmation := ui.NewConfirmation(restoreConfirmationTitle, restoreConfirmationDescription, &confirm)
		confirmation.Run()

		if !confirm {
			fmt.Printf("restore aborted\n")
			return nil
		}

		var backupSelection string = ""
		backupRoot, err := backup.ValidateBackupRootPath(config.BackupRoot)
		if err != nil {
			return err
		}
		directories, err := backup.GetBackupDirectories(backupRoot)
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
