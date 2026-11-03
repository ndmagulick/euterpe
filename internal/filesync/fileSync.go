package filesync

import (
	"crypto/sha256"
	"encoding/hex"
	"euterpe/internal/database"
	"euterpe/internal/filter"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/shirou/gopsutil/v3/disk"
)

type FileData struct {
	Name             string
	Size             int64
	LastModifiedTime time.Time
}

type RecordOperation int

const (
	InsertRecord RecordOperation = iota + 1
	UpdateRecord
)

type CacheUpdate struct {
	FileData  FileData
	Hash      string
	Operation RecordOperation
}

func Sync(source, destination string) error {
	sourceFiles, err := ReadDirectory(source)
	_, err = CheckForStaleFiles(source, sourceFiles)
	if err != nil {
		return err
	}

	var cacheUpdates map[string]CacheUpdate
	cacheUpdates, err = CheckForStaleFiles(source, sourceFiles)
	if err != nil {
		return err
	}

	destinationFiles, err := ReadDirectory(destination)
	if err != nil {
		return err
	}

	filesToDelete, filesToCopy := DiffDirectories(sourceFiles, destinationFiles)
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

	if len(filesToCopy) > 0 || len(cacheUpdates) > 0 {
		log.Printf("Copying files from %s to %s", source, destination)
		errs := copyFilesConcurrent(filesToCopy, source, destination)
		if len(errs) > 0 {
			return fmt.Errorf("%d files failed to copy. See log for details", len(errs))
		}
	}

	return nil
}

func CheckForStaleFiles(path string, files map[string]FileData) (map[string]CacheUpdate, error) {
	cacheUpdates := make(map[string]CacheUpdate)

	for key, value := range files {
		fileHashRecord, err := database.GetFileHashRecord(key)
		if err != nil {
			return nil, err
		}

		if fileHashRecord.Exists && value.LastModifiedTime.UnixNano() == fileHashRecord.LastModifiedTime && value.Size == fileHashRecord.FileSize {
			continue
		}

		var sourceFileHash string
		sourceFileHash, err = hashFile(filepath.Join(path, key))
		if err != nil {
			return nil, err
		}

		if !fileHashRecord.Exists {
			cacheUpdates[key] = CacheUpdate{FileData: value, Hash: sourceFileHash, Operation: InsertRecord}

		} else if sourceFileHash != fileHashRecord.FileHash {
			cacheUpdates[key] = CacheUpdate{FileData: value, Hash: sourceFileHash, Operation: UpdateRecord}
		} else {
			err = database.UpdateFileHashMetadata(key, value.Size, value.LastModifiedTime)

			if err != nil {
				return nil, err
			}
		}
	}

	return cacheUpdates, nil
}

func hashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
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

func DiffDirectories(sourceFiles map[string]FileData, destinationFiles map[string]FileData) (filesToDelete, filesToCopy []FileData) {
	filesToDelete = []FileData{}
	filesToCopy = []FileData{}

	for key, value := range destinationFiles {
		_, ok := sourceFiles[key]
		if !ok {
			filesToDelete = append(filesToDelete, value)
		}
	}

	for key, value := range sourceFiles {
		_, ok := destinationFiles[key]
		if !ok {
			filesToCopy = append(filesToCopy, value)
		}
	}

	return filesToDelete, filesToCopy
}

func calculateSpaceChange(filesToCopy, filesToDelete []FileData) int64 {
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

func deleteFiles(filesToDelete []FileData, path string) error {
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

func copyFilesConcurrent(filesToCopy []FileData, source, destination string) []error {
	const workerCount = 2 // We don't want to overload the bus with too many workers. Anything above 5 may cause problems or even slow down the sync
	jobs := make(chan FileData)
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

func copyFileWorker(jobs <-chan FileData, errors chan<- error, source string, destination string, wg *sync.WaitGroup) {
	defer wg.Done()
	check := color.GreenString("✓")

	for file := range jobs {
		sourceFilePath := filepath.Join(source, file.Name)
		destinationFilePath := filepath.Join(destination, file.Name)
		err := copyFile(sourceFilePath, destinationFilePath)
		if err != nil {
			log.Printf("Failed to copy %s: %v", file.Name, err)
			errors <- err
			continue
		}

		fmt.Printf("%s copied %s\n", check, file.Name)
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

	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return fmt.Errorf("copyFile: %w", err)
	}

	err = destinationFile.Sync()
	if err != nil {
		return fmt.Errorf("copyFile: %w", err)
	}

	return nil
}
