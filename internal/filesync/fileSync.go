package filesync

import (
	"euterpe/internal/filecache"
	"euterpe/internal/filedata"
	"euterpe/internal/filter"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/fatih/color"
	"github.com/shirou/gopsutil/v3/disk"
)

type copyResult struct {
	FileData filedata.FileData
	Err      error
}

func Sync(source, destination string) error {
	sourceFiles, err := ReadDirectory(source)
	var cacheUpdates map[string]filecache.CacheUpdate
	cacheUpdates, err = filecache.CheckForStaleFiles(source, sourceFiles)
	if err != nil {
		return err
	}

	destinationFiles, err := ReadDirectory(destination)
	if err != nil {
		return err
	}

	filesToDelete, filesToCopy := DiffDirectories(sourceFiles, destinationFiles, cacheUpdates)
	syncSize := calculateSpaceChange(filesToCopy, filesToDelete)

	if syncSize > 0 {
		err = checkIfSufficientSpaceExists(destination, syncSize)
		if err != nil {
			return err
		}
	}

	if len(filesToDelete) > 0 {
		log.Printf("Deleting files from %s", destination)
		err := deleteFiles(filesToDelete, destination)
		if err != nil {
			return err
		}
	}

	if len(filesToCopy) > 0 {
		log.Printf("Copying files from %s to %s", source, destination)
		errs := copyFilesConcurrent(filesToCopy, cacheUpdates, source, destination)
		if len(errs) > 0 {
			return fmt.Errorf("%d files failed to copy. See log for details", len(errs))
		}
	}

	return nil
}

func ReadDirectory(path string) (map[string]filedata.FileData, error) {
	directoryEntries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("readDirectory(%s): %w", path, err)
	}

	files := make(map[string]filedata.FileData)

	for _, file := range directoryEntries {
		if file.IsDir() {
			continue
		}

		extension := filepath.Ext(file.Name())
		if !filter.IsAudio(extension) {
			continue
		}

		fileInfo, err := file.Info()
		if err != nil {
			return nil, fmt.Errorf("readDirectory(%s): %w", path, err)
		}

		files[file.Name()] = filedata.FileData{
			Name:             file.Name(),
			Size:             fileInfo.Size(),
			LastModifiedTime: fileInfo.ModTime(),
		}
	}

	return files, nil
}

func DiffDirectories(sourceFiles map[string]filedata.FileData, destinationFiles map[string]filedata.FileData, cacheUpdates map[string]filecache.CacheUpdate) (filesToDelete, filesToCopy []filedata.FileData) {
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

func calculateSpaceChange(filesToCopy, filesToDelete []filedata.FileData) int64 {
	// TODO do we take a space difference for stale files in this calculation?
	var filesToCopyTotalSize int64
	var filesToDeleteTotalSize int64

	for _, data := range filesToCopy {
		filesToCopyTotalSize += data.Size
	}

	for _, data := range filesToDelete {
		filesToDeleteTotalSize += data.Size
	}

	return filesToCopyTotalSize - filesToDeleteTotalSize
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

func deleteFiles(filesToDelete []filedata.FileData, path string) error {
	cross := color.RedString("✕")

	for _, file := range filesToDelete {
		fileToDeletePath := filepath.Join(path, file.Name)
		err := os.Remove(fileToDeletePath)

		if err != nil {
			return fmt.Errorf("deleteFiles: %w", err)
		} else {
			fmt.Printf("%s deleted %s\n", cross, file.Name)
		}
	}

	return nil
}

func copyFilesConcurrent(filesToCopy []filedata.FileData, cacheUpdates map[string]filecache.CacheUpdate, source, destination string) []error {
	const workerCount = 2 // We don't want to overload the bus with too many workers. Anything above 5 may cause problems or even slow down the sync
	const cacheBatchSize = 100
	jobs := make(chan filedata.FileData)
	results := make(chan copyResult)
	var waitGroup sync.WaitGroup

	waitGroup.Add(workerCount)
	for range workerCount {
		go copyFileWorker(jobs, results, source, destination, &waitGroup)
	}

	go func() {
		for _, file := range filesToCopy {
			jobs <- file
		}

		close(jobs)
	}()

	go func() {
		waitGroup.Wait()
		close(results)
	}()

	var errors []error
	batch := make([]filecache.CacheUpdate, 0, cacheBatchSize)

	for result := range results {
		if result.Err != nil {
			errors = append(errors, result.Err)
			continue
		}

		update, ok := cacheUpdates[result.FileData.Name]
		if !ok {
			continue
		}

		batch = append(batch, update)
		if len(batch) >= cacheBatchSize {
			err := flushBatch(batch)
			if err != nil {
				errors = append(errors, err)
			}
			batch = batch[:0]
		}
	}

	// Flush remaining batch after jobs complete
	err := flushBatch(batch)
	if err != nil {
		errors = append(errors, err)
	}

	return errors
}

func copyFileWorker(jobs <-chan filedata.FileData, results chan<- copyResult, source string, destination string, wg *sync.WaitGroup) {
	defer wg.Done()
	check := color.GreenString("✓")

	for file := range jobs {
		sourceFilePath := filepath.Join(source, file.Name)
		destinationFilePath := filepath.Join(destination, file.Name)
		err := copyFile(sourceFilePath, destinationFilePath)
		if err != nil {
			log.Printf("Failed to copy %s: %v", file.Name, err)
		} else {
			fmt.Printf("%s copied %s\n", check, file.Name)
		}

		results <- copyResult{FileData: file, Err: err}
	}
}

func copyFile(source, destination string) error {
	sourceFile, err := os.Open(source)

	if err != nil {
		return fmt.Errorf("copyFile: %w", err)
	}

	defer sourceFile.Close()

	destinationFile, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("copyFile: %w", err)
	}

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return fmt.Errorf("copyFile: %w", err)
	}

	err = destinationFile.Sync()
	if err != nil {
		return fmt.Errorf("copyFile: %w", err)
	}

	err = destinationFile.Close()
	if err != nil {
		os.Remove(destination)
		return fmt.Errorf("copyFile: %w", err)
	}

	return nil
}

func flushBatch(batch []filecache.CacheUpdate) error {
	if len(batch) == 0 {
		return nil
	}

	return nil
}
