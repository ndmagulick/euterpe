package filesync

import (
	"errors"
	"euterpe/internal/directorydiff"
	"euterpe/internal/filecache"
	"euterpe/internal/fileio"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/shirou/gopsutil/v3/disk"
)

type syncPlan struct {
	Inputs directorydiff.DiffInputs
	Diff   directorydiff.DiffResult
}

type SyncReport struct {
	CopyErrors  []error
	CacheErrors []error
}

func Sync(source, destination string) (SyncReport, error) {
	plan, err := planSync(source, destination)
	if err != nil {
		return SyncReport{}, err
	}

	// If there is an error deleting a file, we fail the entire sync since it could cause the rest of the sync to fail due to space constraints on the device
	// Errors copying are collected since a failed copy should not affect other files or space
	copyResults, err := syncFiles(plan.Diff.FilesToDelete, plan.Diff.FilesToCopy, source, destination)
	if err != nil {
		return SyncReport{}, fmt.Errorf("deleting files: %w", err)
	}

	var syncReport SyncReport
	syncReport.CopyErrors = collectCopyErrors(copyResults)
	syncReport.CacheErrors = syncCache(copyResults, plan.Inputs.CacheUpdates, plan.Inputs.SourceFiles)

	if len(syncReport.CopyErrors) > 0 {
		return syncReport, fmt.Errorf("%d of %d files failed to copy: %w", len(syncReport.CopyErrors), len(plan.Diff.FilesToCopy), errors.Join(syncReport.CopyErrors...))
	}

	return syncReport, nil
}

func planSync(source, destination string) (syncPlan, error) {
	diffInputs, err := directorydiff.PrepareDiffInputs(source, destination)
	if err != nil {
		return syncPlan{}, fmt.Errorf("preparing diff: %w", err)
	}

	// If no supported files exist, we error since it would wipe the device and cache
	if len(diffInputs.SourceFiles) == 0 {
		return syncPlan{}, fmt.Errorf("no supported audio files exist in the music library, sync aborted")
	}

	diffResult := directorydiff.Diff(diffInputs.SourceFiles, diffInputs.DestinationFiles, diffInputs.CacheUpdates)
	if diffResult.SyncSize > 0 {
		err = checkIfSufficientSpaceExists(destination, diffResult.SyncSize)
		if err != nil {
			return syncPlan{}, err
		}
	}

	return syncPlan{Inputs: diffInputs, Diff: diffResult}, nil
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
		copySize := calculateCopySize(filesToCopy)
		progressBar := createProgressBar(copySize)
		copyResults = fileio.CopyFilesConcurrent(filesToCopy, source, destination, func(bytesCopied int64) {
			_ = progressBar.Add64(bytesCopied)
		})

		_ = progressBar.Finish()
	}

	return copyResults, nil
}

func collectCopyErrors(copyResults []fileio.CopyResult) []error {
	var copyErrors []error
	for _, result := range copyResults {
		if result.Err != nil {
			copyErrors = append(copyErrors, result.Err)
		}
	}

	return copyErrors
}

func syncCache(copyResults []fileio.CopyResult, cacheUpdates map[string]filecache.CacheUpdate, sourceFiles map[string]fileio.FileData) []error {
	var cacheErrors []error
	const cacheBatchSize = 100
	batch := make([]filecache.CacheUpdate, 0, cacheBatchSize)

	err := filecache.UpdateTouchedFiles(cacheUpdates)
	if err != nil {
		cacheErrors = append(cacheErrors, err)
	}

	for _, result := range copyResults {
		update, ok := cacheUpdates[result.FileData.Name]
		if !ok || result.Err != nil || update.Operation == filecache.UpdateRecordMetadata {
			continue
		}

		batch = append(batch, update)
		if len(batch) >= cacheBatchSize {
			err = filecache.SaveUpdates(batch)
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

	err = filecache.PruneCache(sourceFiles)
	if err != nil {
		cacheErrors = append(cacheErrors, fmt.Errorf("error pruning cache: %w", err))
	}

	return cacheErrors
}

// Progress bar setup functions
func calculateCopySize(filesToCopy []fileio.FileData) int64 {
	var total int64
	for _, fileToCopy := range filesToCopy {
		total += fileToCopy.Size
	}

	return total
}

func createProgressBar(total int64) *progressbar.ProgressBar {
	return progressbar.NewOptions64(total,
		progressbar.OptionSetDescription("copying"),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetElapsedTime(true),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionSetWriter(os.Stderr),
	)
}
