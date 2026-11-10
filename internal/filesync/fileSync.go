package filesync

import (
	"errors"
	"euterpe/internal/directorydiff"
	"euterpe/internal/filecache"
	"euterpe/internal/fileio"
	"fmt"
	"log"

	"github.com/shirou/gopsutil/v3/disk"
)

type SyncReport struct {
	CopyErrors  []error
	CacheErrors []error
}

func Sync(source, destination string) (SyncReport, error) {
	diffInputs, err := directorydiff.PrepareDiffInputs(source, destination)
	if err != nil {
		return SyncReport{}, fmt.Errorf("error preparing diff: %w", err)
	}

	diffResult := directorydiff.Diff(diffInputs.SourceFiles, diffInputs.DestinationFiles, diffInputs.CacheUpdates)
	if diffResult.SyncSize > 0 {
		err = checkIfSufficientSpaceExists(destination, diffResult.SyncSize)
		if err != nil {
			return SyncReport{}, err
		}
	}

	copyResults, deleteErr := syncFiles(diffResult.FilesToDelete, diffResult.FilesToCopy, source, destination)
	if deleteErr != nil {
		return SyncReport{}, fmt.Errorf("error deleting files: %w", deleteErr)
	}

	var syncReport SyncReport
	for _, result := range copyResults {
		if result.Err != nil {
			syncReport.CopyErrors = append(syncReport.CopyErrors, result.Err)
		}
	}

	syncReport.CacheErrors = syncCache(copyResults, diffInputs.CacheUpdates)
	err = filecache.PruneCache(diffInputs.SourceFiles)
	if err != nil {
		syncReport.CacheErrors = append(syncReport.CacheErrors, fmt.Errorf("error pruning cache: %w", err))
	}

	if len(syncReport.CopyErrors) > 0 {
		return syncReport, fmt.Errorf("%d of %d failed to copy: %w", len(syncReport.CopyErrors), len(diffResult.FilesToCopy), errors.Join(syncReport.CopyErrors...))
	}

	return syncReport, nil
}

func checkIfSufficientSpaceExists(destinationPath string, sizeOfSync int64) error {
	usage, err := disk.Usage(destinationPath)
	if err != nil {
		return fmt.Errorf("checkIfSufficientSpaceExists: %w", err)
	}

	if sizeOfSync > int64(usage.Free) {
		return fmt.Errorf("insufficient space. need %d bytes, but only %d bytes available", sizeOfSync, usage.Free)
	}

	return nil
}

func syncFiles(filesToDelete, filesToCopy []fileio.FileData, source, destination string) ([]fileio.CopyResult, error) {
	if len(filesToDelete) > 0 {
		log.Printf("Deleting files from %s", destination)
		err := fileio.DeleteFiles(filesToDelete, destination)
		if err != nil {
			return nil, err
		}
	}

	var copyResults []fileio.CopyResult
	if len(filesToCopy) > 0 {
		log.Printf("Copying files from %s to %s", source, destination)
		copyResults = fileio.CopyFilesConcurrent(filesToCopy, source, destination)
	}

	return copyResults, nil
}

func syncCache(copyResults []fileio.CopyResult, cacheUpdates map[string]filecache.CacheUpdate) []error {
	var cacheErrors []error
	const cacheBatchSize = 100
	batch := make([]filecache.CacheUpdate, 0, cacheBatchSize)

	err := filecache.UpdateTouchedFiles(cacheUpdates)
	if err != nil {
		cacheErrors = append(cacheErrors, err)
	}

	for _, result := range copyResults {
		update, ok := cacheUpdates[result.FileData.Name]
		if !ok || result.Err != nil {
			continue
		}

		batch = append(batch, update)
		if len(batch) >= cacheBatchSize {
			err := filecache.SaveUpdates(batch)
			if err != nil {
				cacheErrors = append(cacheErrors, err)
			}
			batch = batch[:0]
		}
	}

	// Flush remaining batch
	err = filecache.SaveUpdates(batch)
	if err != nil {
		cacheErrors = append(cacheErrors, err)
	}

	return cacheErrors
}
