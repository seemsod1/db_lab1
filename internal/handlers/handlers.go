package handlers

import (
	"fmt"
	"github.com/seemsod1/db_lab1/internal/config"
	"github.com/seemsod1/db_lab1/internal/helpers"
	"github.com/seemsod1/db_lab1/internal/models"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strconv"
)

var Repo *Repository

type Repository struct {
	AppConfig *config.AppConfig
}

func NewRepo(app *config.AppConfig) *Repository {
	return &Repository{AppConfig: app}
}
func NewHandlers(r *Repository) {
	Repo = r
}

func (r *Repository) InsertM(cmd *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <id>: %v\n", err)
		return
	}

	posIn := helpers.GetPosition(uint32(id), r.AppConfig.MasterInd)
	if posIn != -1 {
		fmt.Println("Use another id")
		return
	}

	name := args[1]
	mail := args[2]
	age, err := strconv.Atoi(args[3])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <age>: %v\n", err)
		return
	}
	user := models.User{
		ID:  uint32(id),
		Age: uint32(age),
	}
	copy(user.Name[:], name)
	copy(user.Mail[:], mail)
	user.FirstOrder = -1

	index := models.IndexTable{
		UserId: user.ID,
		Pos:    r.AppConfig.MasterPos,
	}

	if helpers.WriteModel(r.AppConfig.MasterFL, &user, r.AppConfig.MasterPos) {
		r.AppConfig.MasterPos += helpers.MasterSize

		log.Print("User inserted")
		r.AppConfig.MasterInd = append(r.AppConfig.MasterInd, index)

	} else {
		log.Print("User not inserted")
	}
}

func (r *Repository) GetM(cmd *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <id>: %v\n", err)
		return
	}

	pos := helpers.GetPosition(uint32(id), r.AppConfig.MasterInd)
	if pos != -1 {
		var user models.User
		if helpers.ReadModel(r.AppConfig.MasterFL, &user, pos) {
			helpers.PrintMaster(user)
			return
		}

	}

	fmt.Println("User not found")
}

func (r *Repository) GetS(cmd *cobra.Command, args []string) {
	userId, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <user_id>: %v\n", err)
		return
	}

	rentId, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <rent_id>: %v\n", err)
		return
	}

	pos := helpers.GetPosition(uint32(userId), r.AppConfig.SlaveInd)
	if pos == -1 {
		fmt.Println("No orders found for this user")
		return
	}
	readPos := pos
	var tmp models.Order
	for readPos != -1 {
		if !helpers.ReadModel(r.AppConfig.SlaveFL, &tmp, readPos) {
			fmt.Println("Unable to update next_ptr. Error: read failed")
			return
		}

		if tmp.RentId == uint32(rentId) {
			helpers.PrintSlave(tmp)
			return
		}
		readPos = tmp.Next
	}
	fmt.Println("Rent not found")
}

func (r *Repository) InsertS(cmd *cobra.Command, args []string) {
	userId, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <user_id>: %v\n", err)
		return
	}

	isUserExist := helpers.GetPosition(uint32(userId), r.AppConfig.MasterInd)
	if isUserExist == -1 {
		fmt.Println(fmt.Sprintf("Error: User with %d id isn't exist", userId))
		return
	}

	rentId, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <rent_id>: %v\n", err)
		return
	}
	price, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <price>: %v\n", err)
		return
	}

	order := models.Order{
		UserId: uint32(userId),
		RentId: uint32(rentId),
		Price:  uint32(price),
		Next:   -1,
	}
	index := models.IndexTable{
		UserId: order.UserId,
		Pos:    r.AppConfig.SlavePos,
	}
	pos := helpers.GetPosition(uint32(userId), r.AppConfig.SlaveInd)
	if pos == -1 {
		//add first node

		if helpers.AddNode(order, r.AppConfig.SlaveFL, r.AppConfig.SlavePos) {
			r.AppConfig.SlavePos += helpers.SlaveSize
			log.Print("First order inserted")

			r.AppConfig.SlaveInd = append(r.AppConfig.SlaveInd, index)
			return
		} else {
			log.Print("Order not inserted")
			return
		}
	}
	//find last node
	posToInsert := helpers.FindLastNode(r.AppConfig.SlaveFL, pos)
	if posToInsert == -1 {
		fmt.Println("Unable to find last node")
		return
	}

	if helpers.AddNode(order, r.AppConfig.SlaveFL, r.AppConfig.SlavePos, posToInsert) {
		r.AppConfig.SlavePos += helpers.SlaveSize
		log.Print("Order inserted")

	} else {
		log.Print("Order not inserted")
	}

}

func (r *Repository) UtilM(cmd *cobra.Command, args []string) {
	//read all users from master file and print them
	var tmp models.User
	readPos := int64(0) + helpers.MasterSize
	for helpers.ReadModel(r.AppConfig.MasterFL, &tmp, readPos) {
		helpers.PrintMaster(tmp)
		readPos += helpers.MasterSize
	}
	log.Println("All users printed")

}

func (r *Repository) UtilS(cmd *cobra.Command, args []string) {
	//read all orders from slave file and print them
	var tmp models.Order
	readPos := int64(0)
	for helpers.ReadModel(r.AppConfig.SlaveFL, &tmp, readPos) {
		helpers.PrintSlave(tmp)
		readPos += helpers.SlaveSize
	}
	log.Println("All orders printed")
}

func (r *Repository) UpdateM(cmd *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <id>: %v\n", err)
		return
	}

	pos := helpers.GetPosition(uint32(id), r.AppConfig.MasterInd)
	if pos != -1 {
		name := args[1]
		mail := args[2]
		age, err := strconv.Atoi(args[3])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing <age>: %v\n", err)
			return
		}
		user := models.User{
			ID:  uint32(id),
			Age: uint32(age),
		}
		copy(user.Name[:], name)
		copy(user.Mail[:], mail)

		if helpers.WriteModel(r.AppConfig.MasterFL, &user, pos) {
			log.Print("User updated")
			return
		}
	}

	fmt.Println("User not found")
}

func (r *Repository) UpdateS(cmd *cobra.Command, args []string) {
	userId, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <user_id>: %v\n", err)
		return
	}

	rentId, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <rent_id>: %v\n", err)
		return
	}

	pos := helpers.GetPosition(uint32(userId), r.AppConfig.SlaveInd)
	if pos == -1 {
		fmt.Println("No orders found for this user")
		return
	}
	readPos := pos
	var tmp models.Order
	for readPos != -1 {
		if !helpers.ReadModel(r.AppConfig.SlaveFL, &tmp, readPos) {
			fmt.Println("Unable to update next_ptr. Error: read failed")
			return
		}

		if tmp.RentId == uint32(rentId) {
			price, err := strconv.Atoi(args[2])
			if err != nil {
				fmt.Fprintf(os.Stderr, "error parsing <price>: %v\n", err)
				return
			}
			tmp.Price = uint32(price)
			if helpers.WriteModel(r.AppConfig.SlaveFL, &tmp, readPos) {
				log.Print("Order updated")
				return
			}
		}
		readPos = tmp.Next
	}
	fmt.Println("Rent not found")
}
