package folderdiff

import (
	"euterpe/internal/filecache"
	"euterpe/internal/filedata"
)

type DiffResult struct {
	FilesToDelete []filedata.FileData
	FilesToCopy   []filedata.FileData
	CacheUpdates  map[string]filecache.CacheUpdate
	SyncSize      int64
	Err           error
}

func Diff(sourcePath, destinationPath string) DiffResult {
	sourceFiles, err := filedata.ReadDirectory(sourcePath)
	if err != nil {
		return DiffResult{Err: err}
	}

	cacheUpdates, err := filecache.CheckForStaleFiles(sourcePath, sourceFiles)
	if err != nil {
		return DiffResult{Err: err}
	}

	destinationFiles, err := filedata.ReadDirectory(destinationPath)
	if err != nil {
		return DiffResult{Err: err}
	}

	filesToDelete, filesToCopy := diffDirectories(sourceFiles, destinationFiles, cacheUpdates)
	syncSize := calculateSpaceChange(filesToCopy, filesToDelete, destinationFiles)

	return DiffResult{
		FilesToDelete: filesToDelete,
		FilesToCopy:   filesToCopy,
		CacheUpdates:  cacheUpdates,
		SyncSize:      syncSize,
		Err:           nil}
}

func diffDirectories(sourceFiles map[string]filedata.FileData, destinationFiles map[string]filedata.FileData, cacheUpdates map[string]filecache.CacheUpdate) (filesToDelete, filesToCopy []filedata.FileData) {
	filesToDelete = []filedata.FileData{}
	filesToCopy = []filedata.FileData{}

	for key, value := range destinationFiles {
		_, ok := sourceFiles[key]
		if !ok {
			filesToDelete = append(filesToDelete, value)
		}
	}

	for key, value := range sourceFiles {
		_, ok := destinationFiles[key]
		_, needsRecopy := cacheUpdates[key]
		if !ok || needsRecopy {
			filesToCopy = append(filesToCopy, value)
		}
	}

	return filesToDelete, filesToCopy
}

func calculateSpaceChange(filesToCopy, filesToDelete []filedata.FileData, destinationFiles map[string]filedata.FileData) int64 {
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
