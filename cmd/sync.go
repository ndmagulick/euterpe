package cmd

import (
	"euterpe/internal/filesync"
	"euterpe/internal/settings"
	"fmt"

	"github.com/spf13/cobra"
)

var syncCommand = &cobra.Command{
	Use:   "sync",
	Short: "Syncs your music library to your mp3 player",
	Long:  "Syncs the music library specified in your settings.json file to the mp3 player path also specified in your settings.json file.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := settings.ValidateSyncSettings(config.SourceDirectoryPath, config.DestinationDirectoryPath)
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		syncReport, err := filesync.Sync(config.SourceDirectoryPath, config.DestinationDirectoryPath)
		for _, cacheError := range syncReport.CacheErrors {
			fmt.Printf("warning: %v", cacheError)
		}

		if err != nil {
			return err
		}

		return nil
	},
}
