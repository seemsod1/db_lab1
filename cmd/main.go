package main

import (
	"bufio"
	"errors"
	"github.com/seemsod1/db_lab1/internal/config"
	"github.com/seemsod1/db_lab1/internal/driver"
	"github.com/seemsod1/db_lab1/internal/driver/utils"
	myErr "github.com/seemsod1/db_lab1/internal/error"
	"github.com/seemsod1/db_lab1/internal/handlers"
	"log"
	"os"
)

var app config.AppConfig

func main() {

	fileConfig, err := utils.CreateFileConfig(driver.MasterFilename, true)
	if err != nil {
		log.Fatal(err)
	}
	app.Master = fileConfig
	// Open master indexes

	fileConfig, err = utils.CreateFileConfig(driver.SlaveFilename, false)
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

	//close slave file
	app.Slave, err = utils.CloseFile(app.Slave)
	if errors.Is(err, &myErr.Error{Err: myErr.FailedToClose}) {
		log.Fatal("Error: failed to close slave file")
	} else {
		log.Println("Slave file closed")
	}
	//close master file
	app.Master, err = utils.CloseFile(app.Master)
	if errors.Is(err, &myErr.Error{Err: myErr.FailedToClose}) {
		log.Fatal("Error: failed to close master file")
	} else {
		log.Println("Master file closed")
	}

	//save master indexes
	utils.WriteIndices(driver.MasterFilename, app.Master.Ind)
	//save slave indexes
	utils.WriteIndices(driver.SlaveFilename, app.Slave.Ind)
}
