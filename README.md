# Euterpe

![Euterpe](https://upload.wikimedia.org/wikipedia/commons/6/65/Musa3-euterpe-vs.jpg)

This is a personal project meant to make syncing my music library to my mp3 player easier/faster and to learn Go.  
Starting out by just making a simple program, then I want to make a full CLI, and then eventually make a full GUI where you can also play and/or manage music.

### THIS REPOSITORY CONTAINS NO AI GENERATED CODE

## To Run

- You need to have Go on your machine
- Run the makefile
- Replace settings.example.json with settings.json, make sure you have the correct file paths to your library and mp3 player
  - NOTE: Currently, the program will only scan the top level of your music library. Subdirectory support will be added later
- Run the executable

## Things To Do Still

- Make a proper README

#### CLI Functionality (incomplete)

- Print space
- Add command to back up prior to sync
- Add command to restore
- Status bar
- Edit config
- Verify config

#### Much Later (incomplete)

- Add support for expandable storage
- Add support for subdirectories
