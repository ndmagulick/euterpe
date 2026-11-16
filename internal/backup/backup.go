package backup

import (
	"errors"
	"euterpe/internal/fileio"
	"euterpe/ui"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"time"
)

const defaultBackupRootDirectoryName = "Euterpe Backups"
const backupDirectoryNameTimeStamp = "2006-01-02_15-04-05"

func Backup(devicePath, backupRootPath string) (string, error) {
	deviceFileMap, err := fileio.ReadDirectory(devicePath)
	if err != nil {
		return "", err
	}

	if len(deviceFileMap) == 0 {
		return "", fmt.Errorf("device has no audio files to backup")
	}

	backupDirectoryPath, err := createBackupDirectory(backupRootPath)
	if err != nil {
		return "", err
	}

	fmt.Printf("Backing up device files to %s\n", backupDirectoryPath)

	deviceFiles := slices.Collect(maps.Values(deviceFileMap))
	onProgress, finish := ui.NewProgressBar("backing up", fileio.CalculateFileListSize(deviceFiles))
	copyResults := fileio.CopyFilesConcurrent(deviceFiles, devicePath, backupDirectoryPath, onProgress)
	finish()

	copyErrors := fileio.CollectCopyErrors(copyResults)
	if len(copyErrors) > 0 {
		return backupDirectoryPath, fmt.Errorf("%d of %d files failed to backup: %w", len(copyErrors), len(deviceFileMap), errors.Join(copyErrors...))
	}

	return backupDirectoryPath, nil
}

func CreateDefaultBackupRootDirectory() (string, error) {
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

func GetBackupDirectories(backupRootPath string) (map[string]string, error) {
	directoryMap := make(map[string]string)
	entries, err := os.ReadDir(backupRootPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			directoryMap[entry.Name()] = filepath.Join(backupRootPath, entry.Name())
		}
	}

	if len(directoryMap) == 0 {
		return nil, fmt.Errorf("no backups found\n")
	}

	return directoryMap, nil
}

func createBackupDirectory(backupRootPath string) (string, error) {
	backupDirectoryPath := filepath.Join(backupRootPath, time.Now().Format(backupDirectoryNameTimeStamp))
	err := os.MkdirAll(backupRootPath, 0o755)
	if err != nil {
		return "", err
	}

	err = os.Mkdir(backupDirectoryPath, 0o755)
	if err != nil {
		return "", err
	}

	return backupDirectoryPath, nil
}
