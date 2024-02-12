package main

import (
	"fmt"
	"github.com/seemsod1/db_lab1/internal/helpers"
	"github.com/seemsod1/db_lab1/internal/models"
	"log"
	"os"
)

const filename = "user.bin"

func main() {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	user := models.User{}
	helpers.ReadModel(file, &user)
	fmt.Println(user.Name)

	helpers.ReadModel(file, &user)
	fmt.Println(user.Name)

}
