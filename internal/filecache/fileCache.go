package filecache

import (
	"crypto/sha256"
	"encoding/hex"
	"euterpe/internal/database"
	"euterpe/internal/filedata"
	"io"
	"os"
	"path/filepath"
)

type RecordOperation int

const (
	InsertRecord RecordOperation = iota + 1
	UpdateRecord
)

type CacheUpdate struct {
	FileData  filedata.FileData
	Hash      string
	Operation RecordOperation
}

func CheckForStaleFiles(path string, files map[string]filedata.FileData) (map[string]CacheUpdate, error) {
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
