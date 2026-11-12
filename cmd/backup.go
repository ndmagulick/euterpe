package cmd

import (
	"euterpe/internal/backup"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const defaultBackupRootDirectoryName = "Euterpe Backups"

var backupCommand = &cobra.Command{
	Use:   "backup",
	Short: "Backs up your device's audio files to a specified folder",
	Long:  "Backs up your device's audio files to a specified folder under a folder with the timestamp of when the backup took place",
	RunE: func(cmd *cobra.Command, args []string) error {
		backupRootPath, err := validateBackupRootPath(config.BackupRoot)
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

func validateBackupRootPath(backupRootPath string) (string, error) {
	if backupRootPath == "" {
		return createDefaultBackupRootDirectory()
	}

	info, err := os.Stat(backupRootPath)
	if err != nil {
		return "", fmt.Errorf("backup directory %q: %w", backupRootPath, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("back up path %q is not a directory", backupRootPath)
	}

	return backupRootPath, nil
}

func createDefaultBackupRootDirectory() (string, error) {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("accessing home directory to create default backup directory: %w", err)
	}

	defaultRootPath := filepath.Join(homePath, defaultBackupRootDirectoryName)
	err = os.MkdirAll(defaultRootPath, 0o755)
	if err != nil {
		return "", fmt.Errorf("creating default backup directory %q: %w", defaultRootPath, err)
	}

	fmt.Printf("using default backup directory at %s\n", defaultRootPath)

	return defaultRootPath, nil
}
