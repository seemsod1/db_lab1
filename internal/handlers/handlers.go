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

	posIn, _ := helpers.GetPosition(uint32(id), r.AppConfig.MasterInd)
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
		ID:      uint32(id),
		Age:     uint32(age),
		Deleted: false,
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

	pos, _ := helpers.GetPosition(uint32(id), r.AppConfig.MasterInd)
	if pos != -1 {
		var user models.User
		if helpers.ReadModel(r.AppConfig.MasterFL, &user, pos) {
			helpers.PrintMaster(user, false)
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

	pos, _ := helpers.GetPosition(uint32(userId), r.AppConfig.SlaveInd)
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
			fmt.Println("-----------------------------------------------------")
			fmt.Printf("| %-10s | %-10s | %-10s | %-10s |\n", "UserId", "RentId", "Price", "Country")
			fmt.Println("-----------------------------------------------------")
			helpers.PrintSlave(tmp, false)
			fmt.Println("-----------------------------------------------------")
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

	isUserExist, _ := helpers.GetPosition(uint32(userId), r.AppConfig.MasterInd)
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

	country := args[3]

	order := models.Order{
		UserId:  uint32(userId),
		RentId:  uint32(rentId),
		Price:   uint32(price),
		Next:    -1,
		Deleted: false,
	}

	copy(order.Country[:], country)

	index := models.IndexTable{
		UserId: order.UserId,
		Pos:    r.AppConfig.SlavePos,
	}
	pos, _ := helpers.GetPosition(uint32(userId), r.AppConfig.SlaveInd)
	if pos == -1 {
		OrderInsert(r, order)
		r.AppConfig.SlaveInd = append(r.AppConfig.SlaveInd, index)
		return

	}
	//find last node
	lastNodePos := helpers.FindLastNode(r.AppConfig.SlaveFL, pos)
	if lastNodePos == -1 {
		fmt.Println("Unable to find last node")
		return
	}

	OrderInsert(r, order, lastNodePos)

}

func (r *Repository) UtilM(cmd *cobra.Command, args []string) {
	//read all users from master file and print them
	var tmp models.User
	readPos := int64(0) + +helpers.HeaderSize
	for helpers.ReadModel(r.AppConfig.MasterFL, &tmp, readPos) {
		if !tmp.Deleted {
			helpers.PrintMaster(tmp, true)
			readPos += helpers.MasterSize
		}
		readPos += helpers.MasterSize
	}
	log.Println("All users printed")

}

func (r *Repository) UtilS(cmd *cobra.Command, args []string) {
	// read all orders from slave file and print them
	fmt.Println("----------------------------------------------------------------------------")
	fmt.Printf("| %-10s | %-10s | %-10s | %-10s | %-10s | %-7s |\n", "UserId", "RentId", "Price", "Country", "Deleted", "Next")
	fmt.Println("----------------------------------------------------------------------------")
	var gab models.SHeader
	readPos := int64(0)
	helpers.ReadModel(r.AppConfig.SlaveFL, &gab, readPos)
	helpers.PrintGarbage(gab)
	var tmp models.Order
	readPos += helpers.SlaveSize
	fmt.Println("----------------------------------------------------------------------------")
	for helpers.ReadModel(r.AppConfig.SlaveFL, &tmp, readPos) {
		if tmp.Deleted {
			var gab models.SHeader
			helpers.ReadModel(r.AppConfig.SlaveFL, &gab, readPos)
			helpers.PrintGarbage(gab)
			readPos += helpers.SlaveSize
			continue
		}
		helpers.PrintSlave(tmp, true)
		readPos += helpers.SlaveSize
	}
	fmt.Println("----------------------------------------------------------------------------")
	log.Println("All orders printed")
}

func (r *Repository) UpdateM(cmd *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <id>: %v\n", err)
		return
	}

	pos, _ := helpers.GetPosition(uint32(id), r.AppConfig.MasterInd)
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

	pos, _ := helpers.GetPosition(uint32(userId), r.AppConfig.SlaveInd)
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

			country := args[3]

			copy(tmp.Country[:], country)
			if helpers.WriteModel(r.AppConfig.SlaveFL, &tmp, readPos) {
				log.Print("Order updated")
				return
			}
		}
		readPos = tmp.Next
	}
	fmt.Println("Rent not found")
}

func (r *Repository) DeleteS(cmd *cobra.Command, args []string) {
	userId, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <user_id>: %v\n", err)
		return
	}
	rentID, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <rent_id>: %v\n", err)
		return
	}

	pos, index := helpers.GetPosition(uint32(userId), r.AppConfig.SlaveInd)
	if pos == -1 {
		fmt.Println("No user found")
		return
	}
	readPos := pos
	var tmp models.Order
	var prevPos int64
	for readPos != -1 {
		if !helpers.ReadModel(r.AppConfig.SlaveFL, &tmp, readPos) {
			fmt.Println("Unable to update next_ptr. Error: read failed")
			return
		}

		if tmp.RentId == uint32(rentID) {
			if prevPos == 0 {
				if tmp.Next == -1 {
					//only one node in linked list
					OrderDeletion(r, readPos, prevPos)

					//change in index table
					r.AppConfig.SlaveInd = append(r.AppConfig.SlaveInd[:index], r.AppConfig.SlaveInd[index+1:]...)
					return
				}
				//first node in linked list
				OrderDeletion(r, readPos, prevPos)

				//change in index table
				r.AppConfig.SlaveInd[index].Pos = tmp.Next
				return
			}
			//middle or last node in linked list
			OrderDeletion(r, readPos, prevPos)

			return

		}
		prevPos = readPos
		readPos = tmp.Next
	}

	fmt.Println("Order not found")

}

func (r *Repository) CalcS(cmd *cobra.Command, args []string) {
	all, _ := cmd.Flags().GetBool("all")
	var amount uint32
	if all {
		//read all orders from slave file and print them

		var tmp models.Order
		readPos := int64(0)
		for helpers.ReadModel(r.AppConfig.SlaveFL, &tmp, readPos) {
			if tmp.Deleted {
				readPos += helpers.SlaveSize
				continue
			}
			amount++
			readPos += helpers.SlaveSize
		}
		log.Println("Total amount of orders: ", amount)
		if err := cmd.Flags().Set("all", "false"); err != nil {
			log.Println("Error: ", err)
		}
		return
	}
	userId, _ := cmd.Flags().GetInt32("user_id")
	if userId > -1 {
		userId := uint32(userId)

		pos, _ := helpers.GetPosition(userId, r.AppConfig.SlaveInd)
		if pos == -1 {
			log.Println("No user found")
			return
		}
		readPos := pos
		var tmp models.Order
		for readPos != -1 {
			if !helpers.ReadModel(r.AppConfig.SlaveFL, &tmp, readPos) {
				fmt.Println("Unable to update next_ptr. Error: read failed")
				return
			}
			amount++
			readPos = tmp.Next
		}
		log.Println("Total amount of orders: ", amount)
		return
	}
}

func (r *Repository) DeleteM(cmd *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <id>: %v\n", err)
		return
	}

	pos, index := helpers.GetPosition(uint32(id), r.AppConfig.MasterInd)
	if pos != -1 {
		var user models.User
		if helpers.ReadModel(r.AppConfig.MasterFL, &user, pos) {
			user = models.User{Deleted: true}
			if helpers.WriteModel(r.AppConfig.MasterFL, user, pos) {
				//add garbage node

				//delete from index table
				r.AppConfig.MasterInd = append(r.AppConfig.MasterInd[:index], r.AppConfig.MasterInd[index+1:]...)
				log.Print("User deleted")
				return
			}
		}

	}

	fmt.Println("User not found")
}

func OrderDeletion(r *Repository, readPos int64, prev int64) {
	if prev == 0 {
		prev = -1
	}
	helpers.DeleteOrderNode(r.AppConfig.SlaveFL, readPos, prev)
	helpers.AddGarbageNode(r.AppConfig.SlaveFL, r.AppConfig.GarbageSlave, readPos, models.Order{Deleted: true}, r.AppConfig.SlavePos)
	log.Print("Order deleted")
}

func OrderInsert(r *Repository, order models.Order, prev ...int64) {
	var pos int64

	if r.AppConfig.GarbageSlave.Prev == -1 {
		pos = r.AppConfig.GarbageSlave.Next
	} else {
		pos = r.AppConfig.GarbageSlave.Pos
	}
	if len(prev) != 0 {
		helpers.AddNode(order, r.AppConfig.SlaveFL, pos, prev[0])
	} else {
		helpers.AddNode(order, r.AppConfig.SlaveFL, pos)
	}

	if r.AppConfig.GarbageSlave.Prev == -1 {
		r.AppConfig.SlavePos += helpers.SlaveSize
	}
	r.AppConfig.GarbageSlave = helpers.DeleteGarbageNode(r.AppConfig.SlaveFL, r.AppConfig.GarbageSlave, r.AppConfig.GarbageSlave.Prev, r.AppConfig.SlavePos)
	if r.AppConfig.GarbageSlave == nil {
		log.Fatal("Unable to delete node")
	}
	log.Print("Order inserted")
}
