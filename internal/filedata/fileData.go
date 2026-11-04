package filedata

import "time"

type FileData struct {
	Name             string
	Size             int64
	LastModifiedTime time.Time
}
