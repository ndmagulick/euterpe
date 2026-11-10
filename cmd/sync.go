package cmd

import (
	"euterpe/internal/filesync"
	"log"

	"github.com/spf13/cobra"
)

var syncCommand = &cobra.Command{
	Use:   "sync",
	Short: "Syncs your music library to your mp3 player",
	Long:  "Syncs the music library specified in your settings.json file to the mp3 player path also specified in your settings.json file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		syncReport, err := filesync.Sync(config.SourceFilePath, config.DestinationFilePath)

		for _, cacheError := range syncReport.CacheErrors {
			log.Printf("warning: %v", cacheError)
		}

		for _, copyError := range syncReport.CopyErrors {
			log.Printf("copy failed: %v", copyError)
		}

		if err != nil {
			return err
		}

		return nil
	},
}
