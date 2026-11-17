package cmd

import (
	"euterpe/internal/database"
	"euterpe/internal/fileio"
	"euterpe/internal/settings"
	"path/filepath"

	"github.com/spf13/cobra"
)

var config settings.Settings
var executableDirectory string

var rootCommand = &cobra.Command{
	Use:   "euterpe",
	Short: "Euterpe mp3 player sync CLI application",
	Long:  "Euterpe is a simple CLI that syncs your music library to your mp3 player.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		executableDirectory, err = fileio.GetExecutableDirectory()
		if err != nil {
			return err
		}

		cfg, err := settings.ReadSettings(filepath.Join(executableDirectory, settings.SettingsFileName))
		if err != nil {
			return err
		}

		config = cfg
		err = database.OpenDatabase(filepath.Join(executableDirectory, database.DatabaseName))
		if err != nil {
			return err
		}

		return nil
	},
}

func Execute() error {
	return rootCommand.Execute()
}

func init() {
	rootCommand.AddCommand(syncCommand)
	rootCommand.AddCommand(diffCommand)
	rootCommand.AddCommand(versionCommand)
	rootCommand.AddCommand(backupCommand)
	rootCommand.AddCommand(restoreCommand)
}
