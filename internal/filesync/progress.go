package filesync

import (
	"euterpe/internal/fileio"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

func createProgressBar(filesToCopy []fileio.FileData) *progressbar.ProgressBar {
	totalCopySize := calculateCopySize(filesToCopy)

	return progressbar.NewOptions64(totalCopySize,
		progressbar.OptionSetDescription("copying"),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetElapsedTime(true),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionSetWriter(os.Stderr),
	)
}

func calculateCopySize(filesToCopy []fileio.FileData) int64 {
	var total int64
	for _, fileToCopy := range filesToCopy {
		total += fileToCopy.Size
	}

	return total
}
