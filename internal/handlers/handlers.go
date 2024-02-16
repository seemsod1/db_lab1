package handlers

import (
	"fmt"
	"github.com/seemsod1/db_lab1/internal/helpers"
	"github.com/seemsod1/db_lab1/internal/models"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strconv"
	"strings"
)

func (r *Repository) GetM(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "error: 1 argument expected, got %d\n", len(args))
		cmd.Usage()
		return
	}

	var userPosToRead int64
	var all bool

	if args[0] == "all" {
		all = true
	} else {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing ID: %v\n", err)
			return
		}
		if id < 1 {
			fmt.Fprintf(os.Stderr, "error: <id> must be positive and not equal zero\n")
			return
		}

		addressFromIndex, _ := helpers.GetPosition(uint32(id), r.AppConfig.MasterInd)
		if addressFromIndex == -1 {
			fmt.Fprintf(os.Stderr, "record with ID %d not found\n", id)
			return
		}

		userPosToRead = addressFromIndex
	}

	queries := make([]string, 0, len(args)-1)
	for _, q := range args[1:] {
		queries = append(queries, strings.ToLower(q))
	}

	printMasterRecord(r.AppConfig.MasterFL, userPosToRead, queries, all)

}
func (r *Repository) GetS(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "error: 2 argument expected, got %d\n", len(args))
		cmd.Usage()
		return
	}
	var allUsers bool
	var allOrders bool
	var index models.IndexTable
	var rentId int

	if args[0] == "all" {
		allUsers = true
	} else {
		userId, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing <user_id>: %v\n", err)
			return
		}
		if userId < 1 {
			fmt.Fprintf(os.Stderr, "error: <user_id> must be positive and not equal zero\n")
			return
		}

		orderPos, _ := helpers.GetPosition(uint32(userId), r.AppConfig.SlaveInd)
		if orderPos == -1 {
			fmt.Fprintf(os.Stderr, "record with ID %d not found\n", userId)
			return
		}
		index.Pos = orderPos
		index.UserId = uint32(userId)

	}

	if args[1] == "all" {
		allOrders = true
	} else {
		var err error
		rentId, err = strconv.Atoi(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing <rent_id>: %v\n", err)
			return
		}
		if rentId < 1 {
			fmt.Fprintf(os.Stderr, "error: <rent_id> must be positive and not equal zero\n")
			return
		}
	}

	queries := make([]string, 0, len(args)-1)
	for _, q := range args[2:] {
		queries = append(queries, strings.ToLower(q))
	}

	printSlaveRecord(r.AppConfig.SlaveFL, index, uint32(rentId), queries, allUsers, allOrders)
}

func (r *Repository) InsertS(cmd *cobra.Command, args []string) {
	userId, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <user_id>: %v\n", err)
		return
	}
	if userId < 1 {
		fmt.Fprintf(os.Stderr, "error: <user_id> must be positive and not equal zero\n")
		return
	}

	rentId, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <rent_id>: %v\n", err)
		return
	}
	if rentId < 1 {
		fmt.Fprintf(os.Stderr, "error: <rent_id> must be positive and not equal zero\n")
		return
	}
	price, err := strconv.Atoi(args[2])
	if err != nil || price < 1 {
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
		OrderInsert(r, order, index)

		return

	}
	//find last node
	lastNodePos := helpers.FindLastNode(r.AppConfig.SlaveFL, pos)
	if lastNodePos == -1 {
		fmt.Println("Unable to find last node")
		return
	}

	OrderInsert(r, order, models.IndexTable{Pos: -1}, lastNodePos)

}
func (r *Repository) InsertM(cmd *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <id>: %v\n", err)
		return
	}
	if id < 1 {
		fmt.Fprintf(os.Stderr, "error: <id> must be positive and not equal zero\n")
		return
	}

	posIn, _ := helpers.GetPosition(uint32(id), r.AppConfig.MasterInd)
	if posIn != -1 {
		fmt.Println("User already exist")
		return
	}

	name := args[1]
	mail := args[2]
	age, err := strconv.Atoi(args[3])
	if err != nil || age < helpers.MaxAge {
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

	index := models.IndexTable{
		UserId: user.ID,
		Pos:    r.AppConfig.MasterPos,
	}

	UserInsert(r, user, index)
}

func (r *Repository) UtilM(cmd *cobra.Command, args []string) {
	//read all users from master file and print them
	printMasterFile(r.AppConfig.MasterFL)
}
func (r *Repository) UtilS(cmd *cobra.Command, args []string) {
	//read all orders from slave file and print them
	printSlaveFile(r.AppConfig.SlaveFL)
}

func (r *Repository) UpdateM(cmd *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <id>: %v\n", err)
		return
	}

	pos, _ := helpers.GetPosition(uint32(id), r.AppConfig.MasterInd)
	if pos != -1 {
		var originalUser models.User
		if !helpers.ReadModel(r.AppConfig.MasterFL, &originalUser, pos) {
			fmt.Println("Unable to update user. Error: read failed")
			return
		}

		name := args[1]
		if name != "-" {
			copy(originalUser.Name[:], name)
		}
		mail := args[2]
		if mail != "-" {
			copy(originalUser.Mail[:], mail)
		}
		age, err := strconv.Atoi(args[3])
		if err != nil || age < helpers.MaxAge {
			fmt.Fprintf(os.Stderr, "error parsing <age>: %v\n", err)
			return
		}
		originalUser.Age = uint32(age)

		if helpers.WriteModel(r.AppConfig.MasterFL, &originalUser, pos) {
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
			if err != nil || price < 1 {
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
		fmt.Println("User has no orders")
		return
	}
	readPos := pos
	var tmp models.Order
	var prevPos int64
	for readPos != -1 {
		if !helpers.ReadModel(r.AppConfig.SlaveFL, &tmp, readPos) {
			fmt.Println("Failed to delete order from file")
			return
		}

		if tmp.RentId == uint32(rentID) {
			if prevPos == 0 {
				if tmp.Next == -1 {
					//only one node in linked list
					OrderDelete(r, readPos, prevPos)

					//change in index table
					r.AppConfig.SlaveInd = append(r.AppConfig.SlaveInd[:index], r.AppConfig.SlaveInd[index+1:]...)
					return
				}
				//first node in linked list
				OrderDelete(r, readPos, prevPos)

				//change in index table
				r.AppConfig.SlaveInd[index].Pos = tmp.Next
				return
			}
			//middle or last node in linked list
			OrderDelete(r, readPos, prevPos)

			return

		}
		prevPos = readPos
		readPos = tmp.Next
	}

	fmt.Println("Order not found")

}
func (r *Repository) DeleteM(cmd *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <id>: %v\n", err)
		return
	}

	userPos, index := helpers.GetPosition(uint32(id), r.AppConfig.MasterInd)
	headPos, orderIndex := helpers.GetPosition(uint32(id), r.AppConfig.SlaveInd)

	if userPos != -1 {
		var user models.User
		if !helpers.ReadModel(r.AppConfig.MasterFL, &user, userPos) {
			//error handling
			fmt.Println("Unable to delete user. Error: read failed")
			return
		}

		//delete all orders of user
		readPos := headPos
		var tmp models.Order
		for readPos != -1 {
			if !helpers.ReadModel(r.AppConfig.SlaveFL, &tmp, readPos) {
				fmt.Println("Unable to delete user. Error: read failed")
				return
			}
			OrderDelete(r, readPos, 0)
			readPos = tmp.Next
		}

		UserDelete(r, userPos)
		//change in index table
		if headPos != -1 {
			r.AppConfig.SlaveInd = append(r.AppConfig.SlaveInd[:orderIndex], r.AppConfig.SlaveInd[orderIndex+1:]...)
		}

		r.AppConfig.MasterInd = append(r.AppConfig.MasterInd[:index], r.AppConfig.MasterInd[index+1:]...)
		return

	}

	fmt.Println("User not found")
}

func (r *Repository) CalcS(cmd *cobra.Command, args []string) {
	var amount uint32
	if args[0] == "all" {
		//read all orders from slave file and print them
		log.Println("Total amount of orders: ", helpers.NumberOfRecords(r.AppConfig.SlaveInd))
		return

	} else {
		userIdArgs, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing <user_id>: %v\n", err)
			cmd.Usage()
			return
		}
		if userIdArgs < 1 {
			fmt.Fprintf(os.Stderr, "error: <user_id> must be positive and not equal zero\n")
			cmd.Usage()
			return
		}
		userId := uint32(userIdArgs)

		pos, _ := helpers.GetPosition(userId, r.AppConfig.SlaveInd)
		if pos == -1 {
			log.Println("No user found")
			return
		}
		readPos := pos
		var tmp models.Order
		for readPos != -1 {
			if !helpers.ReadModel(r.AppConfig.SlaveFL, &tmp, readPos) {
				fmt.Println("Calculation failed.Error: read failed")
				return
			}
			amount++
			readPos = tmp.Next
		}
		log.Println("Total amount of orders: ", amount)
		return
	}
}
func (r *Repository) CalcM(cmd *cobra.Command, args []string) {
	log.Println("Total amount of users: ", helpers.NumberOfRecords(r.AppConfig.MasterInd))
}
