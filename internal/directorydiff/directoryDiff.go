package directorydiff

import (
	"euterpe/internal/filecache"
	"euterpe/internal/fileio"
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

func compareContents(sourceFiles map[string]fileio.FileData, destinationFiles map[string]fileio.FileData, cacheUpdates map[string]filecache.CacheUpdate) (filesToDelete, filesToCopy []fileio.FileData) {
	filesToDelete = []fileio.FileData{}
	filesToCopy = []fileio.FileData{}

	for key, value := range destinationFiles {
		_, ok := sourceFiles[key]
		if !ok {
			filesToDelete = append(filesToDelete, value)
		}
	}

	for key, value := range sourceFiles {
		destinationFile, onDevice := destinationFiles[key]
		update, existsInCacheUpdates := cacheUpdates[key]
		needsRecopy := existsInCacheUpdates && (update.Operation == filecache.InsertRecord || update.Operation == filecache.UpdateRecordHash)
		sizeMismatch := onDevice && destinationFile.Size != value.Size
		if !onDevice || needsRecopy || sizeMismatch {
			filesToCopy = append(filesToCopy, value)
		}
	}

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
