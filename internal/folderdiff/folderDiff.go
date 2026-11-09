package folderdiff

import (
	"euterpe/internal/filecache"
	"euterpe/internal/filedata"
)

type DiffResult struct {
	FilesToDelete []filedata.FileData
	FilesToCopy   []filedata.FileData
	SyncSize      int64
}

type DiffInputs struct {
	SourceFiles      map[string]filedata.FileData
	DestinationFiles map[string]filedata.FileData
	CacheUpdates     map[string]filecache.CacheUpdate
}

func PrepareDiffInputs(source, destination string) (DiffInputs, error) {
	sourceFiles, err := filedata.ReadDirectory(source)
	if err != nil {
		return DiffInputs{}, err
	}

	destinationFiles, err := filedata.ReadDirectory(destination)
	if err != nil {
		return DiffInputs{}, err
	}

	cacheUpdates, err := filecache.CheckForStaleFiles(source, sourceFiles)
	if err != nil {
		return DiffInputs{}, err
	}

	return DiffInputs{
		SourceFiles:      sourceFiles,
		DestinationFiles: destinationFiles,
		CacheUpdates:     cacheUpdates,
	}, nil
}

func DiffFolderContents(sourceFiles, destinationFiles map[string]filedata.FileData, cacheUpdates map[string]filecache.CacheUpdate) DiffResult {
	filesToDelete, filesToCopy := compareContents(sourceFiles, destinationFiles, cacheUpdates)
	syncSize := calculateSpaceChange(filesToCopy, filesToDelete, destinationFiles)

	return DiffResult{
		FilesToDelete: filesToDelete,
		FilesToCopy:   filesToCopy,
		SyncSize:      syncSize}
}

func compareContents(sourceFiles map[string]filedata.FileData, destinationFiles map[string]filedata.FileData, cacheUpdates map[string]filecache.CacheUpdate) (filesToDelete, filesToCopy []filedata.FileData) {
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
