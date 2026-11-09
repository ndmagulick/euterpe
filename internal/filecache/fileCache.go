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
	UpdateRecordHash
	UpdateRecordMetadata
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
			cacheUpdates[key] = CacheUpdate{FileData: value, Hash: sourceFileHash, Operation: UpdateRecordHash}
		} else {
			cacheUpdates[key] = CacheUpdate{FileData: value, Hash: sourceFileHash, Operation: UpdateRecordMetadata}
		}
	}

	return cacheUpdates, nil
}

func CreateBatch(updates []CacheUpdate) error {
	batch := make([]database.BatchRecord, 0, len(updates))
	for _, update := range updates {
		batch = append(batch, database.BatchRecord{
			Name:         update.FileData.Name,
			Size:         update.FileData.Size,
			ModifiedTime: update.FileData.LastModifiedTime.UnixNano(),
			FileHash:     update.Hash,
			IsInsert:     update.Operation == InsertRecord,
		})
	}

	return database.ApplyBatch(batch)
}

func PruneCache(sourceFiles map[string]filedata.FileData) error {
	filesInCache, err := database.GetAllFileNames()
	if err != nil {
		return err
	}

	orphanedFiles := make([]string, 0)

	for _, file := range filesInCache {
		_, ok := sourceFiles[file]
		if !ok {
			orphanedFiles = append(orphanedFiles, file)
		}
	}

	err = database.DeleteRecords(orphanedFiles)
	if err != nil {
		return err
	}

	return nil
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
