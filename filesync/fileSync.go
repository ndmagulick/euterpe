package filesync

import (
	"euterpe/filter"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
)

type fileData struct {
	name             string
	size             int64
	lastModifiedTime time.Time
}

func Sync(source, destination string) error {
	sourceFiles, err := readDirectory(source)
	if err != nil {
		return err
	}

	destinationFiles, err := readDirectory(destination)
	if err != nil {
		return err
	}

	filesToDelete, filesToCopy := diffDirectories(sourceFiles, destinationFiles)
	syncSize := calculateSpaceChange(filesToCopy, filesToDelete)

	if syncSize > 0 {
		err = checkIfSufficientSpaceExists(destination, syncSize)
		if err != nil {
			return err
		}
	}

	if len(filesToDelete) > 0 {
		err := deleteFiles(filesToDelete, destination)
		if err != nil {
			return err
		}
	}

	if len(filesToCopy) > 0 {
		errs := copyFilesConcurrent(filesToCopy, source, destination)
		if len(errs) > 0 {
			return fmt.Errorf("%d files failed to copy. See log for details.", len(errs))
		}
	}

	return nil
}

func readDirectory(path string) (map[string]fileData, error) {
	directoryEntries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("readDirectory(%s): %w", path, err)
	}

	files := make(map[string]fileData)

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
		files[file.Name()] = fileData{
			name:             file.Name(),
			size:             fileInfo.Size(),
			lastModifiedTime: fileInfo.ModTime(),
		}
	}

	return files, nil
}

func diffDirectories(sourceFiles map[string]fileData, destinationFiles map[string]fileData) (filesToDelete, filesToCopy []fileData) {
	filesToDelete = []fileData{}
	filesToCopy = []fileData{}

	for key, value := range destinationFiles {
		info, ok := sourceFiles[key]
		modTimesAreEqual := timesAreEqualWithTolerance(info.lastModifiedTime, value.lastModifiedTime)
		if !ok || !modTimesAreEqual || info.size != value.size {
			filesToDelete = append(filesToDelete, value)
		}
	}

	for key, value := range sourceFiles {
		info, ok := destinationFiles[key]
		modTimesAreEqual := timesAreEqualWithTolerance(info.lastModifiedTime, value.lastModifiedTime)
		if !ok || !modTimesAreEqual || info.size != value.size {
			filesToCopy = append(filesToCopy, value)
		}
	}

	return filesToDelete, filesToCopy
}

func timesAreEqualWithTolerance(time1, time2 time.Time) bool {
	// most mp3 players use FAT32 which has a 2 second granularity for time
	return time1.Sub(time2).Abs() <= 2*time.Second
}

func calculateSpaceChange(filesToCopy, filesToDelete []fileData) int64 {
	var filesToCopyTotalSize int64
	var filesToDeleteTotalSize int64

	for _, data := range filesToCopy {
		filesToCopyTotalSize += data.size
	}

	for _, data := range filesToDelete {
		filesToDeleteTotalSize += data.size
	}

	return filesToCopyTotalSize - filesToDeleteTotalSize
}

func checkIfSufficientSpaceExists(destinationPath string, sizeOfSync int64) error {
	usage, err := disk.Usage(destinationPath)
	if err != nil {
		return fmt.Errorf("checkIfSufficientSpaceExists: %w", err)
	}

	if sizeOfSync > int64(usage.Free) {
		return fmt.Errorf("insufficient space. need %d bytes, but only %d bytes available.", sizeOfSync, usage.Free)
	}

	return nil
}

func deleteFiles(filesToDelete []fileData, path string) error {
	for _, file := range filesToDelete {
		fileToDeletePath := filepath.Join(path, file.name)
		err := os.Remove(fileToDeletePath)

		if err != nil {
			return fmt.Errorf("deleteFiles: %w", err)
		} else {
			log.Printf("deleted %s from %s", file.name, path)
		}
	}

	return nil
}

func copyFilesConcurrent(filesToCopy []fileData, source, destination string) []error {
	const workerCount = 4 // We don't want to overload the bus with too many workers. Anything above 5 may cause problems or even slow down the sync
	jobs := make(chan fileData)
	errorChan := make(chan error)
	var waitGroup sync.WaitGroup

	waitGroup.Add(workerCount)
	for range workerCount {
		go copyFileWorker(jobs, errorChan, source, destination, &waitGroup)
	}

	for _, file := range filesToCopy {
		jobs <- file
	}

	close(jobs)

	go func() {
		waitGroup.Wait()
		close(errorChan)
	}()

	var errors []error

	for err := range errorChan {
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func copyFileWorker(jobs <-chan fileData, errors chan<- error, source string, destination string, wg *sync.WaitGroup) {
	defer wg.Done()

	for file := range jobs {
		sourceFilePath := filepath.Join(source, file.name)
		destinationFilePath := filepath.Join(destination, file.name)
		err := copyFile(sourceFilePath, destinationFilePath, file.lastModifiedTime)
		if err != nil {
			log.Printf("Failed to copy %s: %v", file.name, err)
			errors <- err
			continue
		}

		log.Printf("copied %s", file.name)
	}
}

func copyFile(source, destination string, sourceModTime time.Time) error {
	sourceFile, err := os.Open(source)

	if err != nil {
		return fmt.Errorf("copyFile: %w", err)
	}

	defer sourceFile.Close()
	destinationFile, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("copyFile: %w", err)
	}

	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return fmt.Errorf("copyFile: %w", err)
	}

	err = destinationFile.Sync()
	if err != nil {
		return fmt.Errorf("copyFile: %w", err)
	}

	// We make the modified time to match the source's since when copying files in Go, it sets it to the same as the creation date
	err = os.Chtimes(destination, time.Time{}, sourceModTime)
	if err != nil {
		return fmt.Errorf("copyFile: %w", err)
	}

	return nil
}
