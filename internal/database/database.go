package database

import (
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

type FileHashRecord struct {
	Exists           bool
	FileSize         int64
	LastModifiedTime int64
	FileHash         string
}

type BatchRecord struct {
	Name         string
	Size         int64
	ModifiedTime int64
	FileHash     string
	IsInsert     bool
}

var db *sql.DB

func OpenDatabase() error {
	var err error
	db, err = sql.Open("sqlite", "cache.db")
	if err != nil {
		return err
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

func CloseDatabase() {
	if db != nil {
		db.Close()
	}
}

func GetFileHashRecord(fileName string) (FileHashRecord, error) {
	var dbFileSize int64
	var dbLastModifiedTime int64
	var dbFileHash string
	query := `SELECT size, lastModifiedTime, fileHash 
			FROM fileHashes 
			WHERE fileName = ?`

	err := db.QueryRow(query, fileName).Scan(&dbFileSize, &dbLastModifiedTime, &dbFileHash)

	if errors.Is(err, sql.ErrNoRows) {
		return FileHashRecord{}, nil
	}

	if err != nil {
		return FileHashRecord{}, err
	}

	return FileHashRecord{Exists: true, FileSize: dbFileSize, LastModifiedTime: dbLastModifiedTime, FileHash: dbFileHash}, nil
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
			_, err = insertStatement.Exec(batchRecord.Name, batchRecord.Size, batchRecord.ModifiedTime, batchRecord.FileHash)
			if err != nil {
				return fmt.Errorf("cache insert failed for file %q: %w", batchRecord.Name, err)
			}
			continue
		}

		var result sql.Result
		result, err = updateStatement.Exec(batchRecord.FileHash, batchRecord.ModifiedTime, batchRecord.Size, batchRecord.Name)
		if err != nil {
			return fmt.Errorf("cache update failed for file %q: %w", batchRecord.Name, err)
		}

		var rowsAffected int64
		rowsAffected, err = result.RowsAffected()
		if err != nil {
			return err
		}

		if rowsAffected == 0 {
			return fmt.Errorf("cache update failed for file %q. No matching rows", batchRecord.Name)
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

	fileNameList := make([]string, 0)

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
