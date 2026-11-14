package restore

import (
	"errors"
	"euterpe/internal/filecache"
	"euterpe/internal/fileio"
	"euterpe/ui"
	"fmt"
	"maps"
	"slices"
)

type restorePlan struct {
	FilesToRestore []fileio.FileData
	FilesToDelete  []fileio.FileData
	RestoreSize    int64
}

func Restore(destination, backupPath string) error {
	plan, err := planRestore(destination, backupPath)
	if err != nil {
		return err
	}

	err = filecache.ClearCache()
	if err != nil {
		return fmt.Errorf("purging cache: %w", err)
	}

	err = fileio.DeleteFiles(plan.FilesToDelete, destination)
	if err != nil {
		return err
	}

	onProgress, finish := ui.NewProgressBar("restoring", plan.RestoreSize)
	copyResults := fileio.CopyFilesConcurrent(plan.FilesToRestore, backupPath, destination, onProgress)
	finish()
	copyErrors := fileio.CollectCopyErrors(copyResults)
	if len(copyErrors) > 0 {
		return fmt.Errorf("%d of %d files failed to copy: %w", len(copyErrors), len(plan.FilesToRestore), errors.Join(copyErrors...))
	}

	return nil
}

func planRestore(destination, backupPath string) (restorePlan, error) {
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
	err = fileio.CheckIfSufficientSpaceExists(destination, sizeDifference)
	if err != nil {
		return restorePlan{}, err
	}

	return restorePlan{
		FilesToRestore: filesToRestore,
		FilesToDelete:  filesToDelete,
		RestoreSize:    backupSize,
	}, nil
}
