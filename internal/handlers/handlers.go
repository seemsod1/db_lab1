package handlers

import (
	"fmt"
	"github.com/seemsod1/db_lab1/internal/driver"
	"github.com/seemsod1/db_lab1/internal/driver/utils"
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
	if len(r.AppConfig.Master.Ind) == 0 {
		fmt.Fprintf(os.Stderr, "error: no records found\n")
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

		addressFromIndex, _ := utils.GetAddressByIndex(uint32(id), r.AppConfig.Master.Ind)
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

	printMasterRecord(r.AppConfig.Master.FL, r.AppConfig.Master.Ind, userPosToRead, queries, all)

}
func (r *Repository) GetS(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "error: 2 argument expected, got %d\n", len(args))
		cmd.Usage()
		return
	}
	if len(r.AppConfig.Slave.Ind) == 0 {
		fmt.Fprintf(os.Stderr, "error: no records found\n")
		return
	}

	var allUsers bool
	var allOrders bool
	var index driver.IndexTable
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

		orderPos, _ := utils.GetAddressByIndex(uint32(userId), r.AppConfig.Slave.Ind)
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

	printSlaveRecord(r.AppConfig.Slave.FL, index, uint32(rentId), queries, allUsers, allOrders)
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
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <price>: %v\n", err)
		return
	}
	if price < 1 {
		fmt.Fprintf(os.Stderr, "error: <price> must be positive and not equal zero\n")
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

	index := driver.IndexTable{
		UserId: order.UserId,
		Pos:    r.AppConfig.Slave.Pos,
	}
	if _, ind := utils.GetAddressByIndex(uint32(userId), r.AppConfig.Master.Ind); ind == -1 {
		fmt.Fprintf(os.Stderr, "error: user with ID %d not found\n", userId)
		return
	}

	pos, _ := utils.GetAddressByIndex(uint32(userId), r.AppConfig.Slave.Ind)
	if pos == -1 {
		OrderInsert(r, order, index)

		return

	}
	//find last node
	lastNodePos := driver.FindLastNode(r.AppConfig.Slave.FL, pos, &models.Order{})
	if lastNodePos == -1 {
		fmt.Fprintf(os.Stderr, "error: last node not found\n")
		return
	}

	OrderInsert(r, order, driver.IndexTable{Pos: -1}, lastNodePos)

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

	posIn, _ := utils.GetAddressByIndex(uint32(id), r.AppConfig.Master.Ind)
	if posIn != -1 {
		fmt.Fprintf(os.Stderr, "error: user with ID %d already exists\n", id)
		return
	}

	name := args[1]
	mail := args[2]
	age, err := strconv.Atoi(args[3])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <age>: %v\n", err)
		return
	}
	if age < driver.MinAge {
		fmt.Fprintf(os.Stderr, "error: <age> must be greater than %d\n", driver.MinAge)
		return
	}
	user := models.User{
		ID:      uint32(id),
		Age:     uint32(age),
		Deleted: false,
	}
	copy(user.Name[:], name)
	copy(user.Mail[:], mail)

	index := driver.IndexTable{
		UserId: user.ID,
		Pos:    r.AppConfig.Master.Pos,
	}

	UserInsert(r, user, index)
}

func (r *Repository) UtilM(cmd *cobra.Command, args []string) {
	//read all users from master file and print them
	printMasterFile(r.AppConfig.Master.FL)
}
func (r *Repository) UtilS(cmd *cobra.Command, args []string) {
	//read all orders from slave file and print them
	printSlaveFile(r.AppConfig.Slave.FL)
}

func (r *Repository) UpdateM(cmd *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <id>: %v\n", err)
		return
	}

	pos, _ := utils.GetAddressByIndex(uint32(id), r.AppConfig.Master.Ind)
	if pos != -1 {
		var originalUser models.User
		if !driver.ReadModel(r.AppConfig.Master.FL, &originalUser, pos) {
			fmt.Fprintf(os.Stderr, "error: unable to update user with ID %d\n", id)
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
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing <age>: %v\n", err)
			return
		}
		if age < driver.MinAge {
			fmt.Fprintf(os.Stderr, "error: <age> must be greater than %d\n", driver.MinAge)
			return
		}
		originalUser.Age = uint32(age)

		if driver.WriteModel(r.AppConfig.Master.FL, &originalUser, pos) {
			log.Print("User updated")
			return
		}
	}

	fmt.Fprintf(os.Stderr, "error: user with ID %d not found\n", id)
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

	pos, _ := utils.GetAddressByIndex(uint32(userId), r.AppConfig.Slave.Ind)
	if pos == -1 {
		fmt.Fprintf(os.Stderr, "error: user with ID %d has no orders\n", userId)
		return
	}
	readPos := pos
	var tmp models.Order
	for readPos != -1 {
		if !driver.ReadModel(r.AppConfig.Slave.FL, &tmp, readPos) {
			fmt.Fprintf(os.Stderr, "error: unable to update order\n")
			return
		}

		if tmp.RentId == uint32(rentId) {
			price, err := strconv.Atoi(args[2])
			if err != nil {
				fmt.Fprintf(os.Stderr, "error parsing <price>: %v\n", err)
				return
			}
			if price < 1 {
				fmt.Fprintf(os.Stderr, "error: <price> must be positive and not equal zero\n")
				return
			}
			tmp.Price = uint32(price)

			country := args[3]

			copy(tmp.Country[:], country)
			if driver.WriteModel(r.AppConfig.Slave.FL, &tmp, readPos) {
				log.Print("Order updated")
				return
			}
		}
		readPos = tmp.Next
	}
	fmt.Fprintf(os.Stderr, "error: order not found\n")
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

	pos, index := utils.GetAddressByIndex(uint32(userId), r.AppConfig.Slave.Ind)
	if pos == -1 {
		fmt.Fprintf(os.Stderr, "error: user with ID %d has no orders\n", userId)
		return
	}
	readPos := pos
	var tmp models.Order
	var prevPos int64
	for readPos != -1 {
		if !driver.ReadModel(r.AppConfig.Slave.FL, &tmp, readPos) {
			fmt.Fprintf(os.Stderr, "error: unable to delete order\n")
			return
		}

		if tmp.RentId == uint32(rentID) {
			if prevPos == 0 {
				if tmp.Next == -1 {
					//only one node in linked list
					OrderDelete(r, readPos, prevPos)

					//change in index table
					r.AppConfig.Slave.Ind = append(r.AppConfig.Slave.Ind[:index], r.AppConfig.Slave.Ind[index+1:]...)
					return
				}
				//first node in linked list
				OrderDelete(r, readPos, prevPos)

				//change in index table
				r.AppConfig.Slave.Ind[index].Pos = tmp.Next
				return
			}
			//middle or last node in linked list
			OrderDelete(r, readPos, prevPos)

			return

		}
		prevPos = readPos
		readPos = tmp.Next
	}

	fmt.Fprintf(os.Stderr, "error: order not found\n")

}
func (r *Repository) DeleteM(cmd *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <id>: %v\n", err)
		return
	}

	userPos, index := utils.GetAddressByIndex(uint32(id), r.AppConfig.Master.Ind)
	headPos, orderIndex := utils.GetAddressByIndex(uint32(id), r.AppConfig.Slave.Ind)

	if userPos != -1 {
		var user models.User
		if !driver.ReadModel(r.AppConfig.Master.FL, &user, userPos) {
			//error handling
			fmt.Fprintf(os.Stderr, "error: unable to delete user with ID %d\n", id)
			return
		}

		//delete all orders of user
		readPos := headPos
		var tmp models.Order
		for readPos != -1 {
			if !driver.ReadModel(r.AppConfig.Slave.FL, &tmp, readPos) {
				fmt.Fprintf(os.Stderr, "error: unable to delete user's order \n")
				return
			}
			OrderDelete(r, readPos, 0)
			readPos = tmp.Next
		}

		UserDelete(r, userPos)
		//change in index table
		if headPos != -1 {
			r.AppConfig.Slave.Ind = append(r.AppConfig.Slave.Ind[:orderIndex], r.AppConfig.Slave.Ind[orderIndex+1:]...)
		}

		r.AppConfig.Master.Ind = append(r.AppConfig.Master.Ind[:index], r.AppConfig.Master.Ind[index+1:]...)
		return

	}

	fmt.Fprintf(os.Stderr, "error: user with ID %d not found\n", id)
}

func (r *Repository) CalcS(cmd *cobra.Command, args []string) {
	var amount uint32
	if args[0] == "all" {
		//read all orders from slave file
		var tmp models.Order
		pos := int64(0)
		for driver.ReadModel(r.AppConfig.Slave.FL, &tmp, pos) {
			if !tmp.Deleted {
				amount++
			}
			pos += driver.OrderSize
		}
		log.Println(fmt.Sprintf("Total amount of orders: %d", amount))
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

		pos, _ := utils.GetAddressByIndex(userId, r.AppConfig.Slave.Ind)
		if pos == -1 {
			fmt.Fprintf(os.Stderr, "error: user with ID %d has no orders\n", userId)
			return
		}
		readPos := pos
		var tmp models.Order
		for readPos != -1 {
			if !driver.ReadModel(r.AppConfig.Slave.FL, &tmp, readPos) {
				fmt.Fprintf(os.Stderr, "error: unable to calculate amount of orders\n")
				return
			}
			amount++
			readPos = tmp.Next
		}
		log.Println(fmt.Sprintf("User with ID %d has %d orders", userId, amount))
		return
	}
}
func (r *Repository) CalcM(cmd *cobra.Command, args []string) {
	log.Println("Total amount of users: ", utils.NumberOfRecords(r.AppConfig.Master.Ind))
}
