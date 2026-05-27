package cmd

import (
	"euterpe/internal/filesync"
	"euterpe/internal/settings"
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
		sourceFiles, err := filesync.ReadDirectory(config.SourceFilePath)
		if err != nil {
			log.Fatal(err)
		}

		destinationFiles, err := filesync.ReadDirectory(config.DestinationFilePath)
		if err != nil {
			log.Fatal(err)
		}

		filesToDelete, filesToCopy := filesync.DiffDirectories(sourceFiles, destinationFiles)

		printDiff(filesToDelete, filesToCopy, config)
	},
}

func printDiff(filesToDelete, filesToCopy []filesync.FileData, config settings.Settings) {
	if len(filesToDelete) > 0 {
		minus := color.RedString("-")
		var sizeToBeFreed int64

		color.Blue("%d files will be deleted from %s", len(filesToDelete), config.DestinationFilePath)
		for _, value := range filesToDelete {
			fmt.Printf("%s %s\n", minus, value.Name)
			sizeToBeFreed += value.Size
		}

		fmt.Printf("\n%d bytes will be freed\n", sizeToBeFreed)
	} else {
		color.Yellow("No files will be deleted")
	}

	if len(filesToCopy) > 0 {
		plus := color.GreenString("+")
		var sizeToBeUsed int64

		color.Blue("%d files will be copied from %s to %s", len(filesToCopy), config.SourceFilePath, config.DestinationFilePath)
		for _, value := range filesToCopy {
			fmt.Printf("%s %s\n", plus, value.Name)
			sizeToBeUsed += value.Size
		}

		fmt.Printf("\n%d bytes will be used\n", sizeToBeUsed)
	} else {
		color.Yellow("No files will be copied")
	}
}
