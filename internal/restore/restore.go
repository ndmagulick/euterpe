package restore

import (
	"errors"
	"euterpe/internal/fileio"
	"euterpe/internal/ui"
	"fmt"
	"maps"
	"os"
	"slices"
)

type restorePlan struct {
	FilesToRestore []fileio.FileData
	FilesToDelete  []fileio.FileData
	RestoreSize    int64
}

func Restore(destination, backupPath string) error {
	_, err := os.Stat(backupPath)
	if err != nil {
		return err
	}
	restorePlan, err := prepareRestore(destination, backupPath)
	if err != nil {
		return err
	}

	err = fileio.DeleteFiles(restorePlan.FilesToDelete, destination)
	if err != nil {
		return err
	}

	onProgress, finish := ui.NewProgressBar("restoring", restorePlan.RestoreSize)
	copyResults := fileio.CopyFilesConcurrent(restorePlan.FilesToRestore, backupPath, destination, onProgress)
	finish()
	copyErrors := fileio.CollectCopyErrors(copyResults)
	if len(copyErrors) > 0 {
		return fmt.Errorf("%d of %d files failed to copy: %w", len(copyErrors), len(restorePlan.FilesToRestore), errors.Join(copyErrors...))
	}

	// invalidate cache entries
	// new function here

	return nil
}

func prepareRestore(destination, backupPath string) (restorePlan, error) {
	backupFileMap, err := fileio.ReadDirectory(backupPath)
	if err != nil {
		return restorePlan{}, err
	}
	filesToRestore := slices.Collect(maps.Values(backupFileMap))

	filesToDeleteMap, err := fileio.ReadDirectory(destination)
	if err != nil {
		return restorePlan{}, err
	}
	filesToDelete := slices.Collect(maps.Values(filesToDeleteMap))

	backupSize := fileio.CalculateFileListSize(filesToRestore)
	filesToDeleteSize := fileio.CalculateFileListSize(filesToDelete)
	sizeDifference := backupSize - filesToDeleteSize
	err = fileio.CheckIfSufficientSpaceExists(destination, sizeDifference) // BUG: bad calculation
	if err != nil {
		return restorePlan{}, err
	}

	return restorePlan{
		FilesToRestore: filesToRestore,
		FilesToDelete:  filesToDelete,
		RestoreSize:    backupSize,
	}, nil
}
