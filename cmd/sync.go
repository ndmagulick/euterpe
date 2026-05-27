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
	Run: func(cmd *cobra.Command, args []string) {
		err := filesync.Sync(config.SourceFilePath, config.DestinationFilePath)

		if err != nil {
			log.Fatal(err)
		}
	},
}
