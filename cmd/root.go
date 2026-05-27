package cmd

import (
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

		return nil
	},
}

func Execute() {
	cobra.CheckErr(rootCommand.Execute())
}

func init() {
	rootCommand.AddCommand(syncCommand)
	rootCommand.AddCommand(diffCommand)
}
