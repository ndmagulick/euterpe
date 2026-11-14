package cmd

import (
	"euterpe/internal/backup"
	"log"

	"github.com/spf13/cobra"
)

// TODO use argument to have custom name
var backupCommand = &cobra.Command{
	Use:   "backup",
	Short: "Backs up your device's audio files to a specified folder",
	Long:  "Backs up your device's audio files to a specified folder under a folder with the timestamp of when the backup took place",
	RunE: func(cmd *cobra.Command, args []string) error {
		backupRootPath, err := backup.ValidateBackupRootPath(config.BackupRoot)
		if err != nil {
			return err
		}

		backupFolderName, err := backup.Backup(config.DestinationFilePath, backupRootPath)
		if err != nil {
			return err
		}

		log.Printf("\nbackup completed, files can be found at %s", backupFolderName)

		return nil
	},
}
