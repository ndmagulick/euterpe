package main

import (
	"euterpe/cmd"
	"euterpe/internal/database"
)

func main() {
	cmd.Execute()
	database.CloseDatabase()
}
