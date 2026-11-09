package filedata

import (
	"euterpe/internal/filter"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type FileData struct {
	Name             string
	Size             int64
	LastModifiedTime time.Time
}

// TODO revisit this, this might be in the wrong place
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
