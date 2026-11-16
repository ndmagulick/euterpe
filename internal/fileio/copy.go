package fileio

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type ProgressFunc func(bytesCopied int64)

type writerOnly struct{ io.Writer }
type progressReader struct {
	io.Reader
	onProgress ProgressFunc
}

type CopyResult struct {
	FileData FileData
	Err      error
}

const bufferSize = 1 << 21 // 2MB buffer

var bufferPool = sync.Pool{
	New: func() any {
		buffer := make([]byte, bufferSize)
		return &buffer
	},
}

func (p progressReader) Read(buffer []byte) (int, error) {
	bytesRead, err := p.Reader.Read(buffer)
	if bytesRead > 0 && p.onProgress != nil {
		p.onProgress(int64(bytesRead))
	}

	return bytesRead, err
}

func CopyFilesConcurrent(filesToCopy []FileData, source, destination string, onProgress ProgressFunc) []CopyResult {
	// We don't want to overload the bus with too many workers. Anything above 5 may cause problems or even slow down the sync
	// NOTE: The bottleneck may reside in the MP3 player depending on its file system and/or how it writes data internally
	const workerCount = 2
	jobs := make(chan FileData)
	workerResults := make(chan CopyResult)
	var finalResults []CopyResult
	var waitGroup sync.WaitGroup

	waitGroup.Add(workerCount)
	for range workerCount {
		go copyFileWorker(jobs, workerResults, source, destination, onProgress, &waitGroup)
	}

	go func() {
		for _, file := range filesToCopy {
			jobs <- file
		}

		close(jobs)
	}()

	go func() {
		waitGroup.Wait()
		close(workerResults)
	}()

	for result := range workerResults {
		finalResults = append(finalResults, result)
	}

	return finalResults
}

func copyFileWorker(jobs <-chan FileData, results chan<- CopyResult, source string, destination string, onProgress ProgressFunc, wg *sync.WaitGroup) {
	defer wg.Done()
	for file := range jobs {
		sourceFilePath := filepath.Join(source, file.Name)
		destinationFilePath := filepath.Join(destination, file.Name)
		err := copyFile(sourceFilePath, destinationFilePath, onProgress)

		results <- CopyResult{FileData: file, Err: err}
	}
}

func copyFile(source, destination string, onProgress ProgressFunc) (err error) {
	bufferPointer := bufferPool.Get().(*[]byte)
	defer bufferPool.Put(bufferPointer)
	sourceFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("copyFile: %w", err)
	}

	defer sourceFile.Close()
	destinationFile, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("copyFile: %w", err)
	}

	defer func() {
		closeErr := destinationFile.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("copyFile %q: %w", destination, closeErr)
		}

		if err != nil {
			_ = os.Remove(destination)
		}
	}()

	_, err = io.CopyBuffer(writerOnly{destinationFile}, progressReader{sourceFile, onProgress}, *bufferPointer)
	if err != nil {
		return fmt.Errorf("copyFile: %w", err)
	}

	// Calls to Sync() cause a slowdown while copying, but ensures data is intact in the event of a power failure or device is removed suddenly
	// Can possibly add a --fast flag to bypass but with warning
	err = destinationFile.Sync()
	if err != nil {
		return fmt.Errorf("copyFile: %w", err)
	}

	return nil
}

func CollectCopyErrors(copyResults []CopyResult) []error {
	var copyErrors []error
	for _, result := range copyResults {
		if result.Err != nil {
			copyErrors = append(copyErrors, result.Err)
		}
	}

	return copyErrors
}
