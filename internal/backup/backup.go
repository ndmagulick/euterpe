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
	"strings"
	"time"
)

const defaultBackupRootDirectoryName = "Euterpe Backups"
const backupDirectoryNameTimeStamp = "2006-01-02_15-04-05.000"
const partialSuffix = ".partial"

func Backup(devicePath, backupRootPath string) (string, error) {
	deviceFileMap, err := fileio.ReadDirectory(devicePath)
	if err != nil {
		return "", err
	}

	if len(deviceFileMap) == 0 {
		return "", fmt.Errorf("device has no audio files to backup")
	}

	deviceFiles := slices.Collect(maps.Values(deviceFileMap))
	backupSize := fileio.CalculateFileListSize(deviceFiles)
	err = fileio.CheckIfSufficientSpaceExists(backupRootPath, backupSize)
	if err != nil {
		return "", err
	}

	backupDirectoryPath, err := createBackupDirectory(backupRootPath)
	if err != nil {
		return "", err
	}

	fmt.Printf("Backing up device files to %s\n", backupDirectoryPath)

	onProgress, finish := ui.NewProgressBar("backing up", backupSize)
	copyResults := fileio.CopyFilesConcurrent(deviceFiles, devicePath, backupDirectoryPath, onProgress)
	finish()

	copyErrors := fileio.CollectCopyErrors(copyResults)
	if len(copyErrors) > 0 {
		copyErrorCollection := fmt.Errorf("%d of %d files failed to backup: %w", len(copyErrors), len(deviceFileMap), errors.Join(copyErrors...))

		var cleanupError error
		removeError := os.RemoveAll(backupDirectoryPath)
		if removeError != nil {
			cleanupError = fmt.Errorf("removing partial backup %q: %w", backupDirectoryPath, removeError)
		}
		return "", errors.Join(copyErrorCollection, cleanupError)
	}

	return commitBackup(backupDirectoryPath)
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

	return defaultRootPath, nil
}

func GetBackupDirectories(backupRootPath string) ([]string, error) {
	var directories []string
	entries, err := os.ReadDir(backupRootPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() && !strings.HasSuffix(entry.Name(), partialSuffix) {
			directories = append(directories, entry.Name())
		}
	}

	if len(directories) == 0 {
		return nil, fmt.Errorf("no backups found in %q", backupRootPath)
	}

	// Sort in descending order with the most recent backup first
	slices.SortFunc(directories, func(a, b string) int {
		return strings.Compare(b, a)
	})

	return directories, nil
}

func createBackupDirectory(backupRootPath string) (string, error) {
	backupDirectoryPath := filepath.Join(backupRootPath, time.Now().Format(backupDirectoryNameTimeStamp)+partialSuffix)
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

func commitBackup(backupDirectoryPath string) (string, error) {
	committedBackupPath := strings.TrimSuffix(backupDirectoryPath, partialSuffix)
	err := os.Rename(backupDirectoryPath, committedBackupPath)
	if err != nil {
		return "", fmt.Errorf("committing backup at %q: %w", backupDirectoryPath, err)
	}

	return committedBackupPath, nil
}
