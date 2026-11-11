package progress

import (
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

func NewProgressBar(description string, totalCopySize int64) (onProgress func(bytesCopied int64), finish func()) {
	bar := progressbar.NewOptions64(totalCopySize,
		progressbar.OptionSetDescription(description),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetElapsedTime(true),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionSetWriter(os.Stderr),
	)

	onProgress = func(bytesCopied int64) {
		_ = bar.Add64(bytesCopied)
	}

	finish = func() {
		_ = bar.Finish()
	}

	return onProgress, finish
}
