package main

import (
	"bufio"
	"github.com/seemsod1/db_lab1/internal/config"
	"github.com/seemsod1/db_lab1/internal/driver"
	"github.com/seemsod1/db_lab1/internal/driver/utils"
	"github.com/seemsod1/db_lab1/internal/handlers"
	"log"
	"os"
)

var app config.AppConfig

func main() {

	fileConfig, err := driver.CreateFileConfig(driver.MasterFilename, true)
	if err != nil {
		log.Fatal(err)
	}
	app.Master = fileConfig
	// Open master indexes

	fileConfig, err = driver.CreateFileConfig(driver.SlaveFilename, false)
	if err != nil {
		log.Fatal(err)
	}
	app.Slave = fileConfig
	// Open slave indexes

	log.Println("Config created")
	log.Println("Starting application")

	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	rootCmd := initRootCommands()
	reader := bufio.NewReader(os.Stdin)

	err = run(rootCmd, reader)
	if err != nil {
		log.Fatal(err)
	}

	//save master indexes
	utils.WriteIndices(driver.MasterFilename, app.Master.Ind)
	//save slave indexes
	utils.WriteIndices(driver.SlaveFilename, app.Slave.Ind)
	//close master file
	if !utils.CloseFile(app.Master, true) {
		log.Fatal("Error: closing master file")
	}
	//close slave file
	if !utils.CloseFile(app.Slave, false) {
		log.Fatal("Error: closing slave file")
	}
}
