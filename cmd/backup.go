package cmd

import (
	"euterpe/internal/backup"
	"log"

	"github.com/spf13/cobra"
)

var backupCommand = &cobra.Command{
	Use:   "backup",
	Short: "Backs up your device's audio files to a specified folder",
	Long:  "Backs up your device's audio files to a specified folder under a folder with the timestamp of when the backup took place",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO add backup path in settings.json, add logic for when a backup path is not specified
		// Add warning for when it's not specified, and copy to a backup directory where the code is
		backupFolderName, err := backup.Backup(config.DestinationFilePath, "")
		if err != nil {
			return err
		}

		log.Printf("Backup completed, files can be found at %s", backupFolderName)

		return nil
	},
}
