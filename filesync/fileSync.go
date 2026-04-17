package filesync

import (
	"euterpe/filter"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// TODO concurrency

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

	if len(filesToDelete) > 0 {
		err := deleteFiles(filesToDelete, destination)
		if err != nil {
			return err
		}
	}

	if len(filesToCopy) > 0 {
		err := copyFiles(filesToCopy, source, destination)
		if err != nil {
			return err
		}
	}

	return nil
}

func readDirectory(path string) (map[string]int64, error) {
	directoryEntries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("readDirectory(%s): %w", path, err)
	}

	files := make(map[string]int64)

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

		files[file.Name()] = fileInfo.Size()
	}

	return files, nil
}

func diffDirectories(sourceFiles map[string]int64, destinationFiles map[string]int64) (filesToDelete, filesToCopy []string) {
	filesToDelete = []string{}
	filesToCopy = []string{}

	// TODO do a better file comparison, size isn't a true indicator of file modification
	for key, value := range destinationFiles {
		size, ok := sourceFiles[key]
		if !ok || size != value {
			filesToDelete = append(filesToDelete, key)
		}
	}

	for key, value := range sourceFiles {
		size, ok := destinationFiles[key]
		if !ok || size != value {
			filesToCopy = append(filesToCopy, key)
		}
	}

	return filesToDelete, filesToCopy
}

func deleteFiles(filesToDelete []string, path string) error {
	for _, file := range filesToDelete {
		fileToDeletePath := filepath.Join(path, file)
		err := os.Remove(fileToDeletePath)

		if err != nil {
			return fmt.Errorf("deleteFiles: %w", err)
		} else {
			log.Printf("deleted %s from %s", file, path)
		}
	}

	return nil
}

func copyFiles(filesToCopy []string, source, destination string) error {
	for _, file := range filesToCopy {
		sourceFilePath := filepath.Join(source, file)
		destinationFilePath := filepath.Join(destination, file)
		err := copyFile(sourceFilePath, destinationFilePath)

		if err != nil {
			return fmt.Errorf("copyFiles: %w", err)
		} else {
			log.Printf("copied %s", file)
		}
	}

	return nil
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

	return destinationFile.Sync()
}
