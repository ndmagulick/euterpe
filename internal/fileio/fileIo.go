package fileio

import (
	"euterpe/internal/filter"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fatih/color"
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

type CopyResult struct {
	FileData FileData
	Err      error
}

type FileData struct {
	Name             string
	Size             int64
	LastModifiedTime time.Time
}

func ReadDirectory(path string) (map[string]FileData, error) {
	directoryEntries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("readDirectory(%s): %w", path, err)
	}

	files := make(map[string]FileData)

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

		files[file.Name()] = FileData{
			Name:             file.Name(),
			Size:             fileInfo.Size(),
			LastModifiedTime: fileInfo.ModTime(),
		}
	}

	return files, nil
}

func DeleteFiles(filesToDelete []FileData, path string) error {
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

func CopyFilesConcurrent(filesToCopy []FileData, source, destination string) []CopyResult {
	// We don't want to overload the bus with too many workers. Anything above 5 may cause problems or even slow down the sync
	// NOTE: The bottleneck may reside in the MP3 player depending on its file system and/or how it writes data internally
	const workerCount = 2
	jobs := make(chan FileData)
	workerResults := make(chan CopyResult)
	finalReults := make([]CopyResult, 0)
	var waitGroup sync.WaitGroup

	waitGroup.Add(workerCount)
	for range workerCount {
		go copyFileWorker(jobs, workerResults, source, destination, &waitGroup)
	}

	go func() {
		for _, file := range filesToCopy {
			jobs <- file
		}

		close(jobs)
	}()

	go func() {
		waitGroup.Wait()
		close(workerResults)
	}()

	for result := range workerResults {
		finalReults = append(finalReults, result)
	}

	return finalReults
}

func copyFileWorker(jobs <-chan FileData, results chan<- CopyResult, source string, destination string, wg *sync.WaitGroup) {
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

		results <- CopyResult{FileData: file, Err: err}
	}
}

func copyFile(source, destination string) (err error) {
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

	defer func() {
		closeErr := destinationFile.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("copyFile %q: %w", destination, closeErr)
		}

		if err != nil {
			_ = os.Remove(destination)
		}
	}()

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

	return nil
}
