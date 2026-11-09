package filesync

import (
	"euterpe/internal/directorydiff"
	"euterpe/internal/filecache"
	"euterpe/internal/filedata"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/fatih/color"
	"github.com/shirou/gopsutil/v3/disk"
)

type readerOnly struct{ io.Reader }
type writerOnly struct{ io.Writer }

const bufferSize = 1 << 21 // 2MB buffer

var bufferPool = sync.Pool{
	New: func() any {
		buffer := make([]byte, bufferSize)
		return &buffer
	},
}

type copyResult struct {
	FileData filedata.FileData
	Err      error
}

func Sync(source, destination string) error {
	diffInputs, err := directorydiff.PrepareDiffInputs(source, destination)
	if err != nil {
		return err
	}

	diffResult := directorydiff.Diff(diffInputs.SourceFiles, diffInputs.DestinationFiles, diffInputs.CacheUpdates)
	if diffResult.SyncSize > 0 {
		err = checkIfSufficientSpaceExists(destination, diffResult.SyncSize)
		if err != nil {
			return err
		}
	}

	if len(diffResult.FilesToDelete) > 0 {
		log.Printf("Deleting files from %s", destination)
		err = deleteFiles(diffResult.FilesToDelete, destination)
		if err != nil {
			return err
		}
	}

	if len(diffResult.FilesToCopy) > 0 {
		log.Printf("Copying files from %s to %s", source, destination)
		errs := copyFilesConcurrent(diffResult.FilesToCopy, diffInputs.CacheUpdates, source, destination)
		if len(errs) > 0 {
			return fmt.Errorf("%d files failed to copy. See log for details", len(errs))
		}
	}

	err = filecache.PruneCache(diffInputs.SourceFiles)
	if err != nil {
		return fmt.Errorf("Error cleaning cache: %w", err)
	}

	return nil
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
	// We don't want to overload the bus with too many workers. Anything above 5 may cause problems or even slow down the sync
	// NOTE: The bottleneck may reside in the MP3 player depending on its file system and/or how it writes data internally
	const workerCount = 2
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
	bufferPointer := bufferPool.Get().(*[]byte)
	defer bufferPool.Put(bufferPointer)

	sourceFile, err := os.Open(source)

	if err != nil {
		return fmt.Errorf("copyFile: %w", err)
	}

	defer sourceFile.Close()

	destinationFile, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("copyFile: %w", err)
	}

	_, err = io.CopyBuffer(writerOnly{destinationFile}, readerOnly{sourceFile}, *bufferPointer)
	if err != nil {
		return fmt.Errorf("copyFile: %w", err)
	}

	// Calls to Sync() cause a slowdown while copying, but ensures data is intact in the event of a power failure or device is removed suddenly
	// Can possibly add a --fast flag to bypass but with warning
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

	err := filecache.CreateBatch(batch)
	if err != nil {
		return err
	}

	return nil
}
