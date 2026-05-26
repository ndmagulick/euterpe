package main

import (
	"euterpe/filesync"
	"euterpe/settings"
	"log"
)

func main() {
	config, err := settings.ReadSettings()
	if err != nil {
		log.Fatal(err)
	}

	err = settings.ValidateSettings(config)
	if err != nil {
		log.Fatal(err)
	}

	err = filesync.Sync(config.SourceFilePath, config.DestinationFilePath)
	if err != nil {
		log.Fatal(err)
	}
}
