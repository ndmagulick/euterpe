package filesync

import (
	"euterpe/internal/directorydiff"
	"euterpe/internal/filecache"
	"euterpe/internal/fileio"
	"fmt"
	"log"

	"github.com/shirou/gopsutil/v3/disk"
)

type SyncErrors struct {
	CopyErrors  []error
	CacheErrors []error
}

// TODO fold errors, this is a desing smell
func Sync(source, destination string) (error, SyncErrors) {
	diffInputs, err := directorydiff.PrepareDiffInputs(source, destination)
	if err != nil {
		return err, SyncErrors{}
	}

	diffResult := directorydiff.Diff(diffInputs.SourceFiles, diffInputs.DestinationFiles, diffInputs.CacheUpdates)
	if diffResult.SyncSize > 0 {
		err = checkIfSufficientSpaceExists(destination, diffResult.SyncSize)
		if err != nil {
			return err, SyncErrors{}
		}
	}

	copyResults, deleteErr := syncFiles(diffResult.FilesToDelete, diffResult.FilesToCopy, source, destination)
	if deleteErr != nil {
		return fmt.Errorf("Error deleting file: %w", deleteErr), SyncErrors{}
	}

	cacheErrors := syncCache(copyResults, diffInputs.CacheUpdates)
	err = filecache.PruneCache(diffInputs.SourceFiles)
	if err != nil {
		cacheErrors = append(cacheErrors, err)
	}

	syncErrors := collectErrors(copyResults, cacheErrors)

	return nil, syncErrors
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

func collectErrors(copyResults []fileio.CopyResult, cacheErrors []error) SyncErrors {
	var syncErrors SyncErrors

	for _, result := range copyResults {
		if result.Err != nil {
			syncErrors.CopyErrors = append(syncErrors.CopyErrors, result.Err)
		}
	}

	syncErrors.CacheErrors = cacheErrors

	return syncErrors
}
