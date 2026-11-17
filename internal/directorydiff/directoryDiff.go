package directorydiff

import (
	"euterpe/internal/filecache"
	"euterpe/internal/fileio"
	"fmt"
	"slices"
	"strings"

	"github.com/fatih/color"
)

type DiffResult struct {
	FilesToDelete []fileio.FileData
	FilesToCopy   []fileio.FileData
	SyncSize      int64
}

type DiffInputs struct {
	SourceFiles      map[string]fileio.FileData
	DestinationFiles map[string]fileio.FileData
	CacheUpdates     map[string]filecache.CacheUpdate
}

func PrepareDiffInputs(source, destination string) (DiffInputs, error) {
	sourceFiles, err := fileio.ReadDirectory(source)
	if err != nil {
		return DiffInputs{}, err
	}

	destinationFiles, err := fileio.ReadDirectory(destination)
	if err != nil {
		return DiffInputs{}, err
	}

	cachedFiles, err := filecache.LoadCachedRecords()
	if err != nil {
		return DiffInputs{}, err
	}

	cacheUpdates, err := filecache.CheckForStaleFiles(source, sourceFiles, cachedFiles)
	if err != nil {
		return DiffInputs{}, err
	}

	return DiffInputs{
		SourceFiles:      sourceFiles,
		DestinationFiles: destinationFiles,
		CacheUpdates:     cacheUpdates,
	}, nil
}

func Diff(sourceFiles, destinationFiles map[string]fileio.FileData, cacheUpdates map[string]filecache.CacheUpdate) DiffResult {
	filesToDelete, filesToCopy := compareContents(sourceFiles, destinationFiles, cacheUpdates)
	syncSize := calculateSpaceChange(filesToCopy, filesToDelete, destinationFiles)

	return DiffResult{
		FilesToDelete: filesToDelete,
		FilesToCopy:   filesToCopy,
		SyncSize:      syncSize}
}

func PrintDiff(diffResult DiffResult, sourcePath, destinationPath string) {
	minus := color.RedString("-")
	plus := color.GreenString("+")

	if len(diffResult.FilesToDelete) > 0 {
		color.Blue("%d files will be deleted from %s", len(diffResult.FilesToDelete), destinationPath)
		for _, value := range diffResult.FilesToDelete {
			fmt.Printf("%s %s\n", minus, value.Name)
		}
	} else {
		color.Yellow("No files will be deleted")
	}

	if len(diffResult.FilesToCopy) > 0 {
		color.Blue("%d files will be copied from %s to %s", len(diffResult.FilesToCopy), sourcePath, destinationPath)
		for _, value := range diffResult.FilesToCopy {
			fmt.Printf("%s %s\n", plus, value.Name)
		}
	} else {
		color.Yellow("No files will be copied")
	}

	if diffResult.SyncSize > 0 {
		fmt.Printf("\n%d bytes will be used\n", diffResult.SyncSize)
	} else if diffResult.SyncSize < 0 {
		fmt.Printf("\n%d bytes will be freed\n", -diffResult.SyncSize)
	} else {
		fmt.Printf("\nNo bytes will be freed or used\n")
	}
}

func compareContents(sourceFiles, destinationFiles map[string]fileio.FileData, cacheUpdates map[string]filecache.CacheUpdate) (filesToDelete, filesToCopy []fileio.FileData) {
	filesToDelete = []fileio.FileData{}
	filesToCopy = []fileio.FileData{}

	for key, value := range destinationFiles {
		_, ok := sourceFiles[key]
		if !ok {
			filesToDelete = append(filesToDelete, value)
		}
	}

	slices.SortFunc(filesToDelete, func(a, b fileio.FileData) int {
		return strings.Compare(a.Name, b.Name)
	})

	for key, value := range sourceFiles {
		destinationFile, onDevice := destinationFiles[key]
		update, existsInCacheUpdates := cacheUpdates[key]
		needsRecopy := existsInCacheUpdates && (update.Operation == filecache.InsertRecord || update.Operation == filecache.UpdateRecordHash)
		sizeMismatch := onDevice && destinationFile.Size != value.Size
		if !onDevice || needsRecopy || sizeMismatch {
			filesToCopy = append(filesToCopy, value)
		}
	}

	slices.SortFunc(filesToCopy, func(a, b fileio.FileData) int {
		return strings.Compare(a.Name, b.Name)
	})

	return filesToDelete, filesToCopy
}

func calculateSpaceChange(filesToCopy, filesToDelete []fileio.FileData, destinationFiles map[string]fileio.FileData) int64 {
	var filesToCopyTotalSize int64
	var filesToDeleteTotalSize int64

	for _, data := range filesToCopy {
		destinationFile, ok := destinationFiles[data.Name]

		// We need to see if the existing file is already on the device and take the difference in case we need to re-copy
		if ok {
			filesToCopyTotalSize += data.Size - destinationFile.Size
		} else {
			filesToCopyTotalSize += data.Size
		}

	}

	for _, data := range filesToDelete {
		filesToDeleteTotalSize += data.Size
	}

	return filesToCopyTotalSize - filesToDeleteTotalSize
}
