package fileio

import (
	"fmt"

	"github.com/shirou/gopsutil/v3/disk"
)

func CalculateFileListSize(fileList []FileData) int64 {
	var total int64
	for _, fileToCopy := range fileList {
		total += fileToCopy.Size
	}

	return total
}

func CheckIfSufficientSpaceExists(destinationPath string, sizeOfChange int64) error {
	usage, err := disk.Usage(destinationPath)
	if err != nil {
		return fmt.Errorf("checkIfSufficientSpaceExists: %w", err)
	}

	if sizeOfChange > int64(usage.Free) {
		return fmt.Errorf("insufficient space at %s: need %d bytes, but only %d bytes available", destinationPath, sizeOfChange, usage.Free)
	}

	return nil
}
