package main

import (
	"euterpe/cmd"
	"euterpe/internal/database"
	"os"
)

func main() {
	err := cmd.Execute()
	err = database.CloseDatabase()

	if err != nil {
		os.Exit(1)
	}
}
