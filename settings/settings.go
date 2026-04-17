package settings

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Settings struct {
	SourceFilePath      string `json:"source"`
	DestinationFilePath string `json:"destination"`
}

func ReadSettings() (settings Settings, err error) {
	file, err := os.Open("settings.json")
	if err != nil {
		return Settings{}, fmt.Errorf("readSettings: %w", err)
	}

	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return Settings{}, fmt.Errorf("readSettings: %w", err)
	}

	err = json.Unmarshal(bytes, &settings)
	if err != nil {
		return Settings{}, fmt.Errorf("readSettings: %w", err)
	}

	return settings, nil
}
