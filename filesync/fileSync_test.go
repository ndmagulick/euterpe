package filesync

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dummyFile struct {
	name    string
	content string
}

func TestReadDirectory(t *testing.T) {
	// Arrange
	testDirectory := t.TempDir()
	fileList := []dummyFile{
		{"mp3test.mp3", "data"},
		{"aacTest.aac", "data"},
		{"m4aTest.m4a", "data"},
		{"flacTest.flac", "data"},
		{"wavTest.wav", "data"},
		{"oggTest.ogg", "data"},
		{"wmaTest.wma", "data"},
		{"alacTest.alac", "data"},
		{"aiffTest.aiff", "data"},
		{"pdfTest.pdf", "data"},
		{"txtTest.txt", "data"},
	}

	err := generateFiles(testDirectory, fileList)
	require.NoError(t, err)

	err = os.Mkdir(filepath.Join(testDirectory, "dir1"), 0755)
	require.NoError(t, err)

	err = os.Mkdir(filepath.Join(testDirectory, "dir2"), 0755)
	require.NoError(t, err)

	// Act
	musicFiles, err := readDirectory(testDirectory)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, 9, len(musicFiles))
	assert.NotContains(t, musicFiles, "pdfTest.pdf")
	assert.NotContains(t, musicFiles, "txtTest.txt")
	assert.NotContains(t, musicFiles, "dir1")
	assert.NotContains(t, musicFiles, "dir2")
}

func TestDiffDirectories(t *testing.T) {
	// Arrange
	dir1 := t.TempDir()
	dir1FileList := []dummyFile{
		{"a.mp3", "data"},
		{"b.mp3", "data"},
		{"c.mp3", "data"},
		{"d.mp3", "data"},
		{"e.mp3", "data"},
		{"f.mp3", "data"},
	}
	err := generateFiles(dir1, dir1FileList)
	require.NoError(t, err)

	dir2 := t.TempDir()
	dir2FileList := []dummyFile{
		{"a.mp3", "modifiedData"},
		{"c.mp3", "data"},
		{"d.mp3", "data"},
		{"x.mp3", "data"},
		{"y.mp3", "data"},
	}
	err = generateFiles(dir2, dir2FileList)
	require.NoError(t, err)

	// Act
	dir1FileData, _ := readDirectory(dir1)
	dir2FileData, _ := readDirectory(dir2)
	filesToDelete, filesToCopy := diffDirectories(dir1FileData, dir2FileData)

	// Assert
	assert.Equal(t, 3, len(filesToDelete))
	assert.True(t, fileDataListContainsFileName("a.mp3", filesToDelete))
	assert.True(t, fileDataListContainsFileName("x.mp3", filesToDelete))
	assert.True(t, fileDataListContainsFileName("y.mp3", filesToDelete))

	assert.Equal(t, 4, len(filesToCopy))
	assert.True(t, fileDataListContainsFileName("a.mp3", filesToCopy))
	assert.True(t, fileDataListContainsFileName("b.mp3", filesToCopy))
	assert.True(t, fileDataListContainsFileName("e.mp3", filesToCopy))
	assert.True(t, fileDataListContainsFileName("f.mp3", filesToCopy))
}

func TestTimesAreEqualWithTolerance(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		testCase string
		time1    time.Time
		time2    time.Time
		expected bool
	}{
		{"equal time", now, now, true},
		{"within tolerance", now, now.Add(1 * time.Second), true},
		{"at tolerance boundary", now, now.Add(2 * time.Second), true},
		{"outside tolerance boundary", now, now.Add(3 * time.Second), false},
		{"negative within boundary", now, now.Add(-1 * time.Second), true},
		{"negative outside boundary", now, now.Add(-3 * time.Second), false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testCase, func(t *testing.T) {
			actual := timesAreEqualWithTolerance(testCase.time1, testCase.time2)
			assert.Equal(t, testCase.expected, actual)
		})
	}
}

func TestCalculateSpaceChange(t *testing.T) {
	testCases := []struct {
		testCase      string
		filesToCopy   []fileData
		filesToDelete []fileData
		expected      int64
	}{
		{"filesToCopy is larger", []fileData{{name: "test.mp3", size: 25, lastModifiedTime: time.Time{}}}, []fileData{{name: "test.wav", size: 10, lastModifiedTime: time.Time{}}}, 15},
		{"filesToDelete is larger", []fileData{{name: "test.mp3", size: 25, lastModifiedTime: time.Time{}}}, []fileData{{name: "test.flac", size: 250, lastModifiedTime: time.Time{}}}, -225},
		{"filesToCopy and filesToDelete are equal", []fileData{{name: "test.mp3", size: 25, lastModifiedTime: time.Time{}}}, []fileData{{name: "test.ogg", size: 25, lastModifiedTime: time.Time{}}}, 0},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testCase, func(t *testing.T) {
			actual := calculateSpaceChange(testCase.filesToCopy, testCase.filesToDelete)
			assert.Equal(t, testCase.expected, actual)
		})
	}
}

func TestSync(t *testing.T) {
	// Arrange
	source := t.TempDir()
	sourceFileList := []dummyFile{
		{"a.mp3", "data"},
		{"b.mp3", "data"},
		{"c.mp3", "data"},
		{"d.mp3", "data"},
		{"e.mp3", "data"},
		{"f.mp3", "data"},
	}
	err := generateFiles(source, sourceFileList)
	require.NoError(t, err)

	destination := t.TempDir()
	destinationFileList := []dummyFile{
		{"a.mp3", "modifiedData"},
		{"c.mp3", "data"},
		{"d.mp3", "data"},
		{"x.mp3", "data"},
		{"y.mp3", "data"},
		{"z.mp3", "data"},
		{"alpha.mp3", "data"},
	}
	err = generateFiles(destination, destinationFileList)
	require.NoError(t, err)

	// Act
	Sync(source, destination)

	// Assert
	actual, err := os.ReadDir(destination)
	require.NoError(t, err)
	assert.Equal(t, 6, len(actual))

	expectedFileNames := []string{"a.mp3", "b.mp3", "c.mp3", "d.mp3", "e.mp3", "f.mp3"}
	for _, name := range expectedFileNames {
		_, err := os.Stat(filepath.Join(destination, name))
		assert.NoError(t, err, "%s should exist in the destination", name)
	}

	expectedDeletedFileNames := []string{"x.mp3", "y.mp3", "z.mp3", "alpha.mp3"}
	for _, name := range expectedDeletedFileNames {
		_, err := os.Stat(filepath.Join(destination, name))
		assert.ErrorIs(t, err, os.ErrNotExist, "%s should have been deleted from destination", name)
	}

	// verify "a.mp3"'s data was overwritten properly
	actualContent, err := os.ReadFile(filepath.Join(destination, "a.mp3"))
	require.NoError(t, err)
	assert.Equal(t, []byte("data"), actualContent)
}

func generateFiles(filePath string, fileList []dummyFile) error {

	for _, file := range fileList {
		err := os.WriteFile(filepath.Join(filePath, file.name), []byte(file.content), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func fileDataListContainsFileName(name string, fileDataList []fileData) bool {
	for _, value := range fileDataList {
		if value.name == name {
			return true
		}
	}

	return false
}
