package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type FileHashRecord struct {
	Exists           bool
	FileSize         int64
	LastModifiedTime int64
	FileHash         string
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

	var count int
	err = db.QueryRow(`SELECT COUNT(*) as count
							FROM sqlite_master 
							WHERE type = 'table'
							AND name = 'fileHashes'`).Scan(&count)

	if err != nil {
		return err
	}

	if count > 0 {
		fmt.Printf("Database is good\n")
	} else {
		fmt.Printf("Database does not exist, generating...\n")

		_, err = db.Exec(`CREATE TABLE fileHashes (
							fileName TEXT PRIMARY KEY,
							size 	 INTEGER NOT NULL,
							lastModifiedTime INTEGER NOT NULL,
							fileHash TEXT NOT NULL
						)`)

		if err != nil {
			return err
		}
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

func InsertFileHashRecord(fileName string, fileSize int64, lastModifiedTime time.Time, fileHash string) error {
	query := `INSERT INTO fileHashes (fileName, size, lastModifiedTime, fileHash)
				VALUES (?, ?, ?, ?)`

	_, err := db.Exec(query, fileName, fileSize, lastModifiedTime.UnixNano(), fileHash)

	return err
}

func UpdateFileHashRecord(fileName string, fileSize int64, lastModifiedTime time.Time, fileHash string) error {
	query := `UPDATE fileHashes 
			SET fileHash = ?, 
			lastModifiedTime = ?,
			size = ?
			WHERE fileName = ?`

	_, err := db.Exec(query, fileHash, lastModifiedTime.UnixNano(), fileSize, fileName)

	return err
}

func UpdateFileHashMetadata(fileName string, fileSize int64, lastModifiedTime time.Time) error {
	query := `UPDATE fileHashes 
				SET lastModifiedTime = ?,
				size = ?
				WHERE fileName = ?`

	_, err := db.Exec(query, lastModifiedTime.UnixNano(), fileSize, fileName)

	return err
}

func DeleteFileHashRecord(fileName string) error {
	query := `DELETE FROM fileHashes WHERE fileName = ?`

	_, err := db.Exec(query, fileName)

	return err
}
