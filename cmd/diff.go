package cmd

import (
	"euterpe/internal/directorydiff"
	"euterpe/internal/settings"

	"github.com/spf13/cobra"
)

var diffCommand = &cobra.Command{
	Use:   "diff",
	Short: "Prints the folder diff between the music library and the mp3 player",
	Long:  "Prints the folder diff between the music library and the mp3 player. It will list which files will be copied to and delete from the mp3 player during the sync process.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return settings.ValidateSyncSettings(config.SourceDirectoryPath, config.DestinationDirectoryPath)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		diffInputs, err := directorydiff.PrepareDiffInputs(config.SourceDirectoryPath, config.DestinationDirectoryPath)
		if err != nil {
			return err
		}

		diffResult := directorydiff.Diff(diffInputs.SourceFiles, diffInputs.DestinationFiles, diffInputs.CacheUpdates)
		directorydiff.PrintDiff(diffResult, config.SourceDirectoryPath, config.DestinationDirectoryPath)

		return nil
	},
}
