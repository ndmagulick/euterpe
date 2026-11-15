package cmd

import (
	"euterpe/internal/backup"
	"euterpe/internal/settings"
	"log"

	"github.com/spf13/cobra"
)

// TODO use argument to have custom name
var backupCommand = &cobra.Command{
	Use:   "backup",
	Short: "Backs up your device's audio files to a specified folder",
	Long:  "Backs up your device's audio files to a specified folder under a folder with the timestamp of when the backup took place",
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
		backupFolderName, err := backup.Backup(config.DestinationDirectoryPath, config.BackupRootDirectoryPath)
		if err != nil {
			return err
		}

		log.Printf("\nbackup completed, files can be found at %s", backupFolderName)

		return nil
	},
}
