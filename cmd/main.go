package main

import (
	"bufio"
	"fmt"
	"github.com/kballard/go-shellquote"
	"github.com/seemsod1/db_lab1/internal/config"
	"github.com/seemsod1/db_lab1/internal/handlers"
	"github.com/seemsod1/db_lab1/internal/helpers"
	"os"
	"strings"
)

var app config.AppConfig

func main() {

	file, pos, gab := helpers.OpenMasterFile("user.fl")
	app.MasterFL = file
	app.MasterPos = pos
	app.GarbageMaster = gab

	file, pos, gab = helpers.OpenSlaveFile("order.fl")
	app.SlaveFL = file
	app.SlavePos = pos
	app.GarbageSlave = gab

	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	rootCmd := initRootCmd()
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Interactive mode. Type 'exit' to quit.")
	for {
		fmt.Print("$ ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading command: %v\n", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "exit" {
			fmt.Println("Exiting...")
			break
		}

		args, err := shellquote.Split(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing command: %v\n", err)
			continue
		}

		rootCmd.SetArgs(args)
		if err := rootCmd.Execute(); err != nil {
			fmt.Fprintf(os.Stderr, "error executing command: %v\n", err)
		}
	}

}

/*
Creating linked list of orders

	var writePos int64 = 0
	var prevPos int64 = -1
	order := models.Order{
		UserId: 1,
		RentId: 1,
		Price:  100,
		Next:   -1,
	}
	helpers.AddNode(order, file, writePos, prevPos)
	prevPos = writePos
	writePos += int64(unsafe.Sizeof(order))

	order = models.Order{
		UserId: 1,
		RentId: 2,
		Price:  200,
		Next:   -1,
	}
	helpers.AddNode(order, file, writePos, prevPos)
	prevPos = writePos
	writePos += int64(unsafe.Sizeof(order))

	order = models.Order{
		UserId: 1,
		RentId: 3,
		Price:  300,
		Next:   -1,
	}
	helpers.AddNode(order, file, writePos, prevPos)
	prevPos = writePos
	writePos += int64(unsafe.Sizeof(order))

	helpers.PrintNodes(file, 0)
*/

/*
Creating list of users
file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var writePos int64 = 0

	user := models.User{ID: 1}
	copy(user.Name[:], "Ivan")
	copy(user.Mail[:], "ivan@example.com")
	user.Age = 20
	user.FirstOrder = -1

	helpers.WriteModel(file, user, writePos)
	writePos += int64(unsafe.Sizeof(user))

	user = models.User{ID: 2}
	copy(user.Name[:], "Vlad")
	copy(user.Mail[:], "vlad@example.com")
	user.Age = 18
	user.FirstOrder = -1

	helpers.WriteModel(file, user, writePos)
	writePos += int64(unsafe.Sizeof(user))

	user = models.User{ID: 3}
	copy(user.Name[:], "Vadim")
	copy(user.Mail[:], "vadim@example.com")
	user.Age = 19
	user.FirstOrder = -1

	helpers.WriteModel(file, user, writePos)
	writePos += int64(unsafe.Sizeof(user))

	user = models.User{ID: 4}
	copy(user.Name[:], "Dima")
	copy(user.Mail[:], "dima@example.com")
	user.Age = 18
	user.FirstOrder = -1

	helpers.WriteModel(file, user, writePos)
	writePos += int64(unsafe.Sizeof(user))

	user = models.User{ID: 5}
	copy(user.Name[:], "Den")
	copy(user.Mail[:], "den@example.com")
	user.Age = 17
	user.FirstOrder = -1

	helpers.WriteModel(file, user, writePos)
	writePos += int64(unsafe.Sizeof(user))
*/
