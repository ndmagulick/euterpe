package main

import (
	"euterpe/cmd"
	"euterpe/internal/database"
	"fmt"
	"os"
)

func main() {
	err := cmd.Execute()
	closeError := database.CloseDatabase()
	if closeError != nil {
		fmt.Fprintln(os.Stderr, "erro closing database: %w", closeError)
	}

	if err != nil {
		os.Exit(1)
	}
}
