package cmd

import (
	"euterpe/internal/database"
	"euterpe/internal/settings"

	"github.com/spf13/cobra"
)

var config settings.Settings

var rootCommand = &cobra.Command{
	Use:   "euterpe",
	Short: "Euterpe mp3 player sync CLI application",
	Long:  "Euterpe is a simple CLI that syncs your music library to your mp3 player.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := settings.ReadSettings()
		if err != nil {
			return err
		}

		err = settings.ValidateSettings(cfg)
		if err != nil {
			return err
		}

		config = cfg

		err = database.OpenDatabase()
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
