package settings

import (
	"encoding/json"
	"euterpe/internal/backup"
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

func ValidateSyncSettings(sourceDirectoryPath, destinationDirectoryPath string) error {
	err := fileio.ValidateDirectoryPath(sourceDirectoryPath)
	if err != nil {
		return err
	}

	err = fileio.ValidateDirectoryPath(destinationDirectoryPath)
	if err != nil {
		return err
	}

	overlapping, err := fileio.PathsOverlap(sourceDirectoryPath, destinationDirectoryPath)
	if err != nil {
		return err
	}

	if overlapping {
		return fmt.Errorf("source and destination paths overlap")
	}

	return nil
}

func ResolveBackupSettings(backupRootDirectoryPath string, destinationDirectoryPath string) (string, error) {
	err := fileio.ValidateDirectoryPath(destinationDirectoryPath)
	if err != nil {
		return "", err
	}

	backupRootPath, err := ensureBackupRoot(backupRootDirectoryPath)
	if err != nil {
		return "", err
	}
	overlapping, err := fileio.PathsOverlap(backupRootPath, destinationDirectoryPath)
	if err != nil {
		return "", err
	}

	if overlapping {
		return "", fmt.Errorf("backup and destination paths overlap")
	}

	return backupRootPath, nil
}

func ensureBackupRoot(backupRootPath string) (string, error) {
	if backupRootPath == "" {
		defaultBackupRootPath, err := backup.CreateDefaultBackupRootDirectory()
		if err != nil {
			return "", err
		}

		return defaultBackupRootPath, nil
	}
	err := fileio.ValidateDirectoryPath(backupRootPath)
	if err != nil {
		return "", err
	}

	return backupRootPath, nil
}

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
