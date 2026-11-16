package settings

import (
	"euterpe/internal/backup"
	"euterpe/internal/fileio"
	"fmt"
)

func ValidateSyncSettings(sourceDirectoryPath, destinationDirectoryPath string) error {
	err := fileio.ValidateDirectoryPath(sourceDirectoryPath)
	if err != nil {
		return err
	}

	err = fileio.ValidateDirectoryPath(destinationDirectoryPath)
	if err != nil {
		return err
	}

	overlapping, err := fileio.PathsOverlap(sourceDirectoryPath, destinationDirectoryPath)
	if err != nil {
		return err
	}

	if overlapping {
		return fmt.Errorf("source and destination paths overlap")
	}

	return nil
}

func ResolveBackupSettings(backupRootDirectoryPath string, destinationDirectoryPath string) (string, error) {
	err := fileio.ValidateDirectoryPath(destinationDirectoryPath)
	if err != nil {
		return "", err
	}

	backupRootPath, err := ensureBackupRoot(backupRootDirectoryPath)
	if err != nil {
		return "", err
	}
	overlapping, err := fileio.PathsOverlap(backupRootPath, destinationDirectoryPath)
	if err != nil {
		return "", err
	}

	if overlapping {
		return "", fmt.Errorf("backup and destination paths overlap")
	}

	return backupRootPath, nil
}

func ensureBackupRoot(backupRootPath string) (string, error) {
	if backupRootPath == "" {
		defaultBackupRootPath, err := backup.CreateDefaultBackupRootDirectory()
		if err != nil {
			return "", err
		}

		return defaultBackupRootPath, nil
	}
	err := fileio.ValidateDirectoryPath(backupRootPath)
	if err != nil {
		return "", err
	}

	return backupRootPath, nil
}
