/*
   TODO near future
   - write tests
   CLI functions
   - Print diff
   - Print settings
   - Print space
   - add option to back up files prior to sync
   - add status bar [====>  ] x%
   TODO much later
   - add support for expandable storage
   - add support for sub directories
*/

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
