package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSettings(t *testing.T) {
	sourceDirectory := t.TempDir()
	destinationDirectory := t.TempDir()

	testCases := []struct {
		testCase        string
		sourcePath      string
		destinationPath string
		expectError     bool
	}{
		{"both paths valid", sourceDirectory, destinationDirectory, false},
		{"source path invalid", "\\path\\to\\garbage", destinationDirectory, true},
		{"destination path invalid", sourceDirectory, "\\path\\to\\garbage", true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testCase, func(t *testing.T) {
			actual := ValidateSettings(Settings{SourceFilePath: testCase.sourcePath, DestinationFilePath: testCase.destinationPath})
			if testCase.expectError {
				assert.Error(t, actual)
			} else {
				assert.NoError(t, actual)
			}
		})
	}
}
