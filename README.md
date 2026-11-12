# Euterpe

![Euterpe](https://upload.wikimedia.org/wikipedia/commons/6/65/Musa3-euterpe-vs.jpg)

This is a personal project meant to make syncing my music library to my mp3 player easier/faster and to learn Go.
Still a work in progress :)
<br/>

- Developed on Windows
- Local mp3 player: Sony Walkman NW-A45 with 64gb expandable storage

### THIS REPOSITORY CONTAINS NO AI GENERATED CODE

Claude was only used for review and to be a rubber ducky.

## To Run

- You need to have Go on your machine
- Run the makefile
- Replace settings.example.json with settings.json, make sure you have the correct file paths to your library and mp3 player
  - NOTE: Currently, the program will only scan the top level of your music library. Subdirectory support will be added later
- Run the executable

## Things To Do Still

- Fix the tests
- Make a proper README
- Print space
- Add command to restore
- Edit config via command line
- Verify config via command line

#### Much Later (incomplete)

- Add support for expandable storage
- Add support for subdirectories
