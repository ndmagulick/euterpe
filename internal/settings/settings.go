package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const SettingsFileName = "settings.json"

type Settings struct {
	SourceDirectoryPath      string `json:"source"`
	DestinationDirectoryPath string `json:"destination"`
	BackupRootDirectoryPath  string `json:"backupRoot"`
}

func ReadSettings(settingsPath string) (Settings, error) {
	contentBytes, err := os.ReadFile(settingsPath)
	if err != nil {
		return Settings{}, fmt.Errorf("readSettings %q: %w", settingsPath, err)
	}

	var settings Settings
	err = json.Unmarshal(contentBytes, &settings)
	if err != nil {
		return Settings{}, fmt.Errorf("readSettings %q: %w", settingsPath, err)
	}

	return settings, nil
}

// TODO expand this for when we allow updating of settings via command line
func UpdateBackupPathSetting(executableDirectoryPath, backupPath string) error {
	fullSettingsPath := filepath.Join(executableDirectoryPath, SettingsFileName)
	settings, err := ReadSettings(fullSettingsPath)
	if err != nil {
		return fmt.Errorf("updateBackupPath: %w", err)
	}

	settings.BackupRootDirectoryPath = backupPath
	err = writeSettings(fullSettingsPath, settings)
	if err != nil {
		return fmt.Errorf("updateBackupPath: %w", err)
	}

	return nil
}

func writeSettings(settingsPath string, settings Settings) error {
	data, err := json.MarshalIndent(settings, "", "    ")
	if err != nil {
		return fmt.Errorf("writeSettings: %w", err)
	}

	err = os.WriteFile(settingsPath, data, 0o644)
	if err != nil {
		return fmt.Errorf("writeSettings: %w", err)
	}

	return nil
}
