package settings

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Settings struct {
	SourceDirectoryPath      string `json:"source"`
	DestinationDirectoryPath string `json:"destination"`
	BackupRootDirectoryPath  string `json:"backupRoot"`
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

// TODO expand this for when we allow updating of settings via command line
func UpdateBackupPathSetting(backupPath string) error {
	settings, err := ReadSettings()
	if err != nil {
		return fmt.Errorf("updateBackupPath: %w", err)
	}

	settings.BackupRootDirectoryPath = backupPath
	err = writeSettings(settings)
	if err != nil {
		return fmt.Errorf("updateBackupPath: %w", err)
	}

	return nil
}

func writeSettings(settings Settings) error {
	data, err := json.MarshalIndent(settings, "", "    ")
	if err != nil {
		return fmt.Errorf("writeSettings: %w", err)
	}

	err = os.WriteFile("settings.json", data, 0o644)
	if err != nil {
		return fmt.Errorf("writeSettings: %w", err)
	}

	return nil
}
