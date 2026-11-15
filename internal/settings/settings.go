package settings

import (
	"encoding/json"
	"euterpe/internal/fileio"
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

func ValidateSettings(settings Settings) error {
	err := validateSettingsPath(settings.SourceDirectoryPath)
	if err != nil {
		return err
	}

	err = validateSettingsPath(settings.DestinationDirectoryPath)
	if err != nil {
		return err
	}

	overlapping, err := fileio.PathsOverlap(settings.SourceDirectoryPath, settings.DestinationDirectoryPath)
	if err != nil {
		return err
	}

	if overlapping {
		return fmt.Errorf("source and destination paths overlap")
	}

	return nil
}

func validateSettingsPath(path string) error {
	pathInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("validateSettings: %w", err)
	}

	if !pathInfo.IsDir() {
		return fmt.Errorf("validateSettings: path is not a directory: %q", path)
	}

	return nil
}
