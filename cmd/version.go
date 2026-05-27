package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var versionCommand = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(`
 _______          _________ _______  _______  _______  _______ 
(  ____ \|\     /|\__   __/(  ____ \(  ____ )(  ____ )(  ____ \
| (    \/| )   ( |   ) (   | (    \/| (    )|| (    )|| (    \/
| (__    | |   | |   | |   | (__    | (____)|| (____)|| (__    
|  __)   | |   | |   | |   |  __)   |     __)|  _____)|  __)   
| (      | |   | |   | |   | (      | (\ (   | (      | (      
| (____/\| (___) |   | |   | (____/\| ) \ \__| )      | (____/\
(_______/(_______)   )_(   (_______/|/   \__/|/       (_______/
                                                               
		`)
		color.Green("Euterpe v0.1")
		color.Green("Coded by a human in Berlin")
	},
}
