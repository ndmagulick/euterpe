package cmd

import (
	"errors"
	"euterpe/internal/backup"
	"euterpe/internal/restore"
	"euterpe/internal/settings"
	"euterpe/ui"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

const backupMenuTitle = "Available backups"
const restoreConfirmationTitle = "Are you sure you want to restore?"
const restoreConfirmationDescription = "Restoring from a backup will wipe all audio files currently on your device, and copy all files in the backup folder.\nThis operation will also purge the cache which will slow down your next sync."

var restoreCommand = &cobra.Command{
	Use:   "restore",
	Short: "Restores your device's music library to a previous state from a backup folder",
	Long:  "",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		backupRoot, err := settings.ResolveBackupSettings(config.BackupRootDirectoryPath, config.DestinationDirectoryPath)
		if err != nil {
			return err
		}

		if backupRoot != config.BackupRootDirectoryPath {
			config.BackupRootDirectoryPath = backupRoot
			err = settings.UpdateBackupPathSetting(backupRoot)
			if err != nil {
				return err
			}
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var confirm bool
		confirmation := ui.NewConfirmation(restoreConfirmationTitle, restoreConfirmationDescription, &confirm)
		err := confirmation.Run()
		if err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return nil
			}

			return err
		}

		if !confirm {
			fmt.Printf("restore aborted\n")
			return nil
		}

		var backupSelection string = ""
		directories, err := backup.GetBackupDirectories(config.BackupRootDirectoryPath)
		if err != nil {
			return err
		}

		backupMenu, err := ui.NewSelectMenu(backupMenuTitle, directories, &backupSelection)
		if err != nil {
			return err
		}
		err = backupMenu.Run()
		if err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return nil
			}

			return err
		}

		err = restore.Restore(config.DestinationDirectoryPath, backupSelection)

		return nil
	},
}
