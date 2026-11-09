package cmd

import (
	"euterpe/internal/filedata"
	"euterpe/internal/folderdiff"
	"fmt"
	"log"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var diffCommand = &cobra.Command{
	Use:   "diff",
	Short: "Prints the folder diff between the music library and the mp3 player",
	Long:  "Prints the folder diff between the music library and the mp3 player. It will list which files will be copied to and delete from the mp3 player during the sync process.",
	Run: func(cmd *cobra.Command, args []string) {
		diffInputs, err := folderdiff.PrepareDiffInputs(config.SourceFilePath, config.DestinationFilePath)
		if err != nil {
			log.Fatal(err)
		}

		diffResult := folderdiff.Diff(diffInputs.SourceFiles, diffInputs.DestinationFiles, diffInputs.CacheUpdates)
		printDiff(diffResult.FilesToDelete, diffResult.FilesToCopy, diffResult.SyncSize)
	},
}

func printDiff(filesToDelete, filesToCopy []filedata.FileData, syncSize int64) {
	minus := color.RedString("-")
	plus := color.GreenString("+")

	if len(filesToDelete) > 0 {
		color.Blue("%d files will be deleted from %s", len(filesToDelete), config.DestinationFilePath)
		for _, value := range filesToDelete {
			fmt.Printf("%s %s\n", minus, value.Name)
		}
	} else {
		color.Yellow("No files will be deleted")
	}

	if len(filesToCopy) > 0 {
		color.Blue("%d files will be copied from %s to %s", len(filesToCopy), config.SourceFilePath, config.DestinationFilePath)
		for _, value := range filesToCopy {
			fmt.Printf("%s %s\n", plus, value.Name)
		}
	} else {
		color.Yellow("No files will be copied")
	}

	if syncSize > 0 {
		fmt.Printf("\n%d bytes will be used\n", syncSize)
	} else if syncSize < 0 {
		fmt.Printf("\n%d bytes will be freed\n", syncSize)
	} else {
		fmt.Printf("\nNo bytes will be freed or used\n")
	}
}
