package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func ValidateDatabase() error {
	db, err := sql.Open("sqlite", "cache.db")
	if err != nil {
		return err
	}
	defer db.Close()

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

	return err
}
