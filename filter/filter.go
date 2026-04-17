package filter

import "strings"

var audioExtensions = map[string]struct{}{
	".mp3":  {},
	".aac":  {},
	".m4a":  {},
	".flac": {},
	".wav":  {},
	".ogg":  {},
	".wma":  {},
	".alac": {},
	".aiff": {},
}

func IsAudio(fileExtension string) bool {
	_, ok := audioExtensions[strings.ToLower(fileExtension)]
	return ok
}
