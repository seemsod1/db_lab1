package main

import (
	"bufio"
	"github.com/seemsod1/db_lab1/internal/config"
	"github.com/seemsod1/db_lab1/internal/handlers"
	"github.com/seemsod1/db_lab1/internal/helpers"
	"log"
	"os"
)

var app config.AppConfig

func main() {

	file, pos, gab := helpers.OpenMasterFile("user")
	app.MasterFL = file
	app.MasterPos = pos
	app.GarbageMaster = gab

	file, pos, gab = helpers.OpenSlaveFile("order")
	app.SlaveFL = file
	app.SlavePos = pos
	app.GarbageSlave = gab

	rootCmd := initRootCommands()
	reader := bufio.NewReader(os.Stdin)

	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	err := run(rootCmd, reader)
	if err != nil {
		log.Fatal(err)
	}

}
