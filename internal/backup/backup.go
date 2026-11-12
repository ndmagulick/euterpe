package backup

import (
	"errors"
	"euterpe/internal/fileio"
	"euterpe/internal/progress"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"time"
)

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
	onProgress, finish := progress.NewProgressBar("backing up", fileio.CalculateCopySize(deviceFiles))
	copyResults := fileio.CopyFilesConcurrent(deviceFiles, devicePath, backupDirectoryPath, onProgress)
	finish()

	copyErrors := fileio.CollectCopyErrors(copyResults)
	if len(copyErrors) > 0 {
		return backupDirectoryPath, fmt.Errorf("%d of %d files failed to backup: %w", len(copyErrors), len(deviceFileMap), errors.Join(copyErrors...))
	}

	return backupDirectoryPath, nil
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
