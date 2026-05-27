package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAudio(t *testing.T) {
	testCases := []struct {
		testCase  string
		extension string
		expected  bool
	}{
		{"mp3", ".mp3", true},
		{"mp3 uppercase", ".MP3", true},
		{"aac", ".aac", true},
		{"aac partial uppercase", ".aAC", true},
		{"m4a", ".m4a", true},
		{"flac", ".flac", true},
		{"wav", ".wav", true},
		{"ogg", ".ogg", true},
		{"wma", ".wma", true},
		{"alac", ".alac", true},
		{"aiff", ".aiff", true},
		{"txt", ".txt", false},
		{"pdf uppercase", ".PDF", false},
		{"bin", ".bin", false},
		{"png", ".png", false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testCase, func(t *testing.T) {
			actual := IsAudio(testCase.extension)
			assert.Equal(t, testCase.expected, actual)
		})
	}
}
