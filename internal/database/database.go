package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

const DatabaseName = "cache.db"

type FileHashRecord struct {
	FileName         string
	FileSize         int64
	LastModifiedTime int64
	FileHash         string
}

type BatchRecord struct {
	Data     FileHashRecord
	IsInsert bool
}

var db *sql.DB

func OpenDatabase(databasePath string) error {
	var err error
	db, err = sql.Open("sqlite", databasePath)
	if err != nil {
		return fmt.Errorf("opening cache database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS fileHashes (
							fileName TEXT PRIMARY KEY,
							size 	 INTEGER NOT NULL,
							lastModifiedTime INTEGER NOT NULL,
							fileHash TEXT NOT NULL
						)`)

	if err != nil {
		return err
	}

	return nil
}

func CloseDatabase() error {
	if db != nil {
		err := db.Close()
		if err != nil {
			return fmt.Errorf("closing database: %w", err)
		}
	}

	return nil
}

func ApplyBatch(batchRecords []BatchRecord) error {
	transaction, err := db.Begin()
	if err != nil {
		return err
	}
	defer transaction.Rollback()

	insertStatement, err := transaction.Prepare(`INSERT INTO fileHashes (fileName, size, lastModifiedTime, fileHash)
				VALUES (?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer insertStatement.Close()

	updateStatement, err := transaction.Prepare(`UPDATE fileHashes 
			SET fileHash = ?, 
			lastModifiedTime = ?,
			size = ?
			WHERE fileName = ?`)
	if err != nil {
		return err
	}
	defer updateStatement.Close()

	for _, batchRecord := range batchRecords {
		if batchRecord.IsInsert {
			_, err = insertStatement.Exec(batchRecord.Data.FileName, batchRecord.Data.FileSize, batchRecord.Data.LastModifiedTime, batchRecord.Data.FileHash)
			if err != nil {
				return fmt.Errorf("cache insert failed for file %q: %w", batchRecord.Data.FileName, err)
			}
			continue
		}

		var result sql.Result
		result, err = updateStatement.Exec(batchRecord.Data.FileHash, batchRecord.Data.LastModifiedTime, batchRecord.Data.FileSize, batchRecord.Data.FileName)
		if err != nil {
			return fmt.Errorf("cache update failed for file %q: %w", batchRecord.Data.FileName, err)
		}

		var rowsAffected int64
		rowsAffected, err = result.RowsAffected()
		if err != nil {
			return err
		}

		if rowsAffected == 0 {
			return fmt.Errorf("cache update failed for file %q. No matching rows", batchRecord.Data.FileName)
		}
	}

	return transaction.Commit()
}

func GetAllFileNames() ([]string, error) {
	query := `SELECT fileName FROM fileHashes`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fileNameList []string
	for rows.Next() {
		var fileName string
		err = rows.Scan(&fileName)
		if err != nil {
			return nil, err
		}

		fileNameList = append(fileNameList, fileName)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return fileNameList, err
}

func GetAllFiles() (map[string]FileHashRecord, error) {
	query := `SELECT fileName, size, lastModifiedTime, fileHash
			  FROM fileHashes`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	filesInCache := make(map[string]FileHashRecord)

	for rows.Next() {
		var dbFileName string
		var dbFileSize int64
		var dbLastModifiedTime int64
		var dbFileHash string
		err = rows.Scan(&dbFileName, &dbFileSize, &dbLastModifiedTime, &dbFileHash)
		if err != nil {
			return nil, err
		}

		filesInCache[dbFileName] = FileHashRecord{FileName: dbFileName, FileSize: dbFileSize, LastModifiedTime: dbLastModifiedTime, FileHash: dbFileHash}
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return filesInCache, err
}

func DeleteRecords(orphanedFiles []string) error {
	if len(orphanedFiles) == 0 {
		return nil
	}

	transaction, err := db.Begin()
	if err != nil {
		return err
	}
	defer transaction.Rollback()

	deleteStatement, err := transaction.Prepare(`DELETE FROM fileHashes WHERE fileName = ?`)
	if err != nil {
		return err
	}
	defer deleteStatement.Close()

	for _, orphanedFile := range orphanedFiles {
		_, err = deleteStatement.Exec(orphanedFile)
		if err != nil {
			return fmt.Errorf("cache pruning failed for file %q: %w", orphanedFile, err)
		}
	}

	return transaction.Commit()
}

func ClearTable() error {
	transaction, err := db.Begin()
	if err != nil {
		return err
	}
	defer transaction.Rollback()

	_, err = transaction.Exec(`DELETE FROM fileHashes`)
	if err != nil {
		return err
	}

	return transaction.Commit()
}
