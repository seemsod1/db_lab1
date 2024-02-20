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

		addressFromId := utils.GetAddressById(uint32(id), r.AppConfig.Master.Ind)
		if addressFromId == -1 {
			fmt.Fprintf(os.Stderr, "record with ID %d not found\n", id)
			return
		}

		userPosToRead = addressFromId
	}

	queries := make([]string, 0, len(args)-1)
	for _, q := range args[1:] {
		queries = append(queries, strings.ToLower(q))
	}

	printMasterRecord(r.AppConfig.Master.FL, r.AppConfig.Master.Ind, userPosToRead, queries, all)

}
func (r *Repository) GetS(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Printf("error: at least 1 argument is required, got %d\n", len(args))
		err := cmd.Usage()
		if err != nil {
			return
		}
		return
	}

	if len(r.AppConfig.Slave.Ind) == 0 {
		fmt.Fprintf(os.Stderr, "error: no records found\n")
		return
	}
	orderId := -1
	var all = args[0] == "all"
	var err error

	if !all {
		orderId, err = strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing <order_id>: %v\n", err)
			return
		}
		if orderId < 1 {
			fmt.Fprintf(os.Stderr, "error: <order_id> must be positive and not equal zero\n")
			return
		}
		address := utils.GetAddressById(uint32(orderId), r.AppConfig.Slave.Ind)
		if address == -1 {
			fmt.Printf("record with ID %d not found\n", orderId)
			return
		}
	}

	var userID int
	var fsAddress int64 = -1
	var queries []string

	queries = make([]string, 0, len(args)-1)
	for _, q := range args[1:] {
		queries = append(queries, strings.ToUpper(q))
	}

	if len(args) > 1 {

		userID, err = strconv.Atoi(queries[0])
		if err != nil {
			userID = -1
		} else {

			address := utils.GetAddressById(uint32(userID), r.AppConfig.Master.Ind)
			if address == -1 {
				fmt.Printf("record with ID %d not found\n", userID)
				return
			}

			var model models.User
			if !driver.ReadModel(r.AppConfig.Master.FL, &model, address) {
				fmt.Printf("error reading master data: %s\n", err)
				return
			}

			if model.FirstOrder == -1 {
				fmt.Fprintf(os.Stderr, "user with ID %d has no orders\n", userID)
				return
			}
			fsAddress = model.FirstOrder

		}
	}

	printSlaveRecord(r.AppConfig.Slave.FL, r.AppConfig.Slave.Ind, orderId, fsAddress, queries, all)
}

func (r *Repository) InsertS(_ *cobra.Command, args []string) {
	orderId, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <rent_id>: %v\n", err)
		return
	}
	if orderId < 1 {
		fmt.Fprintf(os.Stderr, "error: <orderId> must be positive and not equal zero\n")
		return
	}
	userId, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <user_id>: %v\n", err)
		return
	}
	if userId < 1 {
		fmt.Fprintf(os.Stderr, "error: <user_id> must be positive and not equal zero\n")
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
	if price > driver.MaxPrice {
		fmt.Fprintf(os.Stderr, "error: <price> must be less than 1mil\n")
		return
	}

	country := args[3]

	order := models.Order{
		OrderId: uint32(orderId),
		UserId:  uint32(userId),
		Price:   uint32(price),
		Next:    -1,
		Prev:    -1,
		Deleted: false,
	}

	copy(order.Country[:], country)

	index := driver.IndexTable{
		Id:  order.OrderId,
		Pos: -1,
	}

	//check if user exists
	userPos := utils.GetAddressById(uint32(userId), r.AppConfig.Master.Ind)
	if userPos == -1 {
		fmt.Fprintf(os.Stderr, "error: user with ID %d not found\n", userId)
		return
	}

	//check if order exists
	pos := utils.GetAddressById(uint32(orderId), r.AppConfig.Slave.Ind)
	if pos != -1 {
		fmt.Fprintf(os.Stderr, "error: such record already exist\n")
		return
	}

	//check if user has orders
	var user models.User
	if !driver.ReadModel(r.AppConfig.Master.FL, &user, userPos) {
		fmt.Fprintf(os.Stderr, "error: unable to insert order\n")
		return
	}
	pos = user.FirstOrder

	if user.FirstOrder == -1 {
		orderInsert(r, order, index, user.ID, true)
		return

	}
	//find last node
	lastNodePos := utils.FindLastNode(r.AppConfig.Slave.FL, pos, &models.Order{})
	if lastNodePos == -1 {
		fmt.Fprintf(os.Stderr, "error: last node not found\n")
		return
	}

	orderInsert(r, order, index, user.ID, false, lastNodePos)

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

	posIn := utils.GetAddressById(uint32(id), r.AppConfig.Master.Ind)
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
	if age > driver.MaxAge {
		fmt.Fprintf(os.Stderr, "error: <age> must be less than %d\n", driver.MaxAge)
		return
	}
	user := models.User{
		ID:         uint32(id),
		Age:        uint32(age),
		Deleted:    false,
		FirstOrder: -1,
	}
	copy(user.Name[:], name)
	copy(user.Mail[:], mail)

	index := driver.IndexTable{
		Id:  user.ID,
		Pos: r.AppConfig.Master.Pos,
	}

	userInsert(r, user, index)
}

func (r *Repository) UtilM(_ *cobra.Command, _ []string) {
	//read all users from master file and print them
	printMasterFile(r.AppConfig.Master.FL)
}
func (r *Repository) UtilS(_ *cobra.Command, _ []string) {
	//read all orders from slave file and print them
	printSlaveFile(r.AppConfig.Slave.FL)
}

func (r *Repository) UpdateM(_ *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <id>: %v\n", err)
		return
	}

	pos := utils.GetAddressById(uint32(id), r.AppConfig.Master.Ind)
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

		arg := args[3]
		if arg != "-" {
			age, err := strconv.Atoi(args[3])
			if err != nil {
				fmt.Fprintf(os.Stderr, "error parsing <age>: %v\n", err)
				return
			}
			if age < driver.MinAge {
				fmt.Fprintf(os.Stderr, "error: <age> must be greater than %d\n", driver.MinAge)
				return
			}
			if age > driver.MaxAge {
				fmt.Fprintf(os.Stderr, "error: <age> must be less than %d\n", driver.MaxAge)
				return
			}
			originalUser.Age = uint32(age)
		}

		if driver.WriteModel(r.AppConfig.Master.FL, &originalUser, pos) {
			log.Print("User updated")
			return
		}
	}

	fmt.Fprintf(os.Stderr, "error: user with ID %d not found\n", id)
}
func (r *Repository) UpdateS(_ *cobra.Command, args []string) {
	orderId, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <order_id>: %v\n", err)
		return
	}

	orderPos := utils.GetAddressById(uint32(orderId), r.AppConfig.Slave.Ind)
	if orderPos == -1 {
		var order models.Order
		if !driver.ReadModel(r.AppConfig.Slave.FL, &order, orderPos) {
			fmt.Fprintf(os.Stderr, "error: unable to update record")
			return
		}

		arg := args[1]
		var price int
		if arg != "-" {
			price, err = strconv.Atoi(args[1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "error parsing <price>: %v\n", err)
				return
			}
			if price < 1 {
				fmt.Fprintf(os.Stderr, "error: <price> must be positive and not equal zero\n")
				return
			}
			if price > driver.MaxPrice {
				fmt.Fprintf(os.Stderr, "error: <price> must be less than 1mil\n")
				return
			}
			order.Price = uint32(price)
		}

		country := args[2]
		if country != "-" {
			copy(order.Country[:], country)
		}
		if driver.WriteModel(r.AppConfig.Slave.FL, &order, orderPos) {
			log.Print("Order updated")
			return
		}

	}
	fmt.Fprintf(os.Stderr, "error: order with ID %d not found\n", orderId)

}

func (r *Repository) DeleteS(_ *cobra.Command, args []string) {
	orderId, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <orderId>: %v\n", err)
		return
	}
	userId, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <user_id>: %v\n", err)
		return
	}

	userPos := utils.GetAddressById(uint32(userId), r.AppConfig.Master.Ind)
	if userPos == -1 {
		fmt.Fprintf(os.Stderr, "error: user with ID %d doesnt exist\n", userId)
		return
	}
	var user models.User
	if !driver.ReadModel(r.AppConfig.Master.FL, &user, userPos) {
		fmt.Fprintf(os.Stderr, "error: unable to delete record")
		return
	}
	if user.FirstOrder == -1 {
		fmt.Fprintf(os.Stderr, "error: user with ID %d has no orders\n", userId)
		return
	}

	orderPos := utils.GetAddressById(uint32(orderId), r.AppConfig.Slave.Ind)
	if orderPos == -1 {
		fmt.Fprintf(os.Stderr, "error: order with ID %d not found\n", orderId)
		return
	}

	readPos := orderPos
	var order models.Order
	prevPos := int64(-1)
	if !driver.ReadModel(r.AppConfig.Slave.FL, &order, readPos) {
		fmt.Fprintf(os.Stderr, "error: unable to delete record")
		return
	}
	if order.UserId != uint32(userId) {
		fmt.Fprintf(os.Stderr, "error: order with ID %d doesnt belong to user with ID %d\n", orderId, userId)
		return
	}
	orderIndex := utils.GetIdByAddress(orderPos, r.AppConfig.Slave.Ind)
	if orderIndex == 0 {
		fmt.Fprintf(os.Stderr, "error: unable to delete order with ID %d\n", orderId)
		return
	}
	if user.FirstOrder == orderPos {
		if order.Next == -1 {
			changeMasterFirstOrder(r.AppConfig.Master.FL, userPos, prevPos)
		} else {
			changeMasterFirstOrder(r.AppConfig.Master.FL, userPos, order.Next)
		}
		orderDelete(r, readPos, prevPos)
	} else {
		//prevPos = utils.FindPrevNode(r.AppConfig.Slave.FL, user.FirstOrder, orderPos, &models.Order{})
		orderDelete(r, readPos, order.Prev)
	}
	r.AppConfig.Slave.Ind = removeById(orderIndex, r.AppConfig.Slave.Ind)

	r.AppConfig.Slave, err = utils.OptimizeFile(r.AppConfig.Slave, r.AppConfig.Master, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: unable to optimize file\n")
		return
	}

}
func (r *Repository) DeleteM(_ *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing <id>: %v\n", err)
		return
	}

	userPos := utils.GetAddressById(uint32(id), r.AppConfig.Master.Ind)
	index := utils.GetIdByAddress(userPos, r.AppConfig.Master.Ind)

	if userPos != -1 {
		var user models.User
		if !driver.ReadModel(r.AppConfig.Master.FL, &user, userPos) {
			//error handling
			fmt.Fprintf(os.Stderr, "error: unable to delete user with ID %d\n", id)
			return
		}

		//delete all orders of user
		readPos := user.FirstOrder
		if readPos != -1 {
			var tmp models.Order
			for readPos != -1 {
				if !driver.ReadModel(r.AppConfig.Slave.FL, &tmp, readPos) {
					fmt.Fprintf(os.Stderr, "error: unable to delete user's order \n")
					return
				}
				orderDelete(r, readPos, -1)
				orderIndex := utils.GetIdByAddress(readPos, r.AppConfig.Slave.Ind)
				r.AppConfig.Slave.Ind = removeById(orderIndex, r.AppConfig.Slave.Ind)
				readPos = tmp.Next
			}
			r.AppConfig.Slave, err = utils.OptimizeFile(r.AppConfig.Slave, r.AppConfig.Master, false)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: unable to optimize file\n")
				return
			}
		}

		userDelete(r, userPos)
		r.AppConfig.Master.Ind = removeById(index, r.AppConfig.Master.Ind)
		//change in index table

		r.AppConfig.Master, err = utils.OptimizeFile(r.AppConfig.Master, nil, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: unable to optimize file\n")
			return
		}
		return

	}

	fmt.Fprintf(os.Stderr, "error: user with ID %d not found\n", id)
}

func (r *Repository) CalcS(cmd *cobra.Command, args []string) {
	var amount int
	if args[0] == "all" {
		//read all orders from slave file
		log.Println(fmt.Sprintf("Total amount of orders: %d", numberOfRecords(r.AppConfig.Slave.Ind)))
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
		userPos := utils.GetAddressById(userId, r.AppConfig.Master.Ind)
		if userPos == -1 {
			fmt.Fprintf(os.Stderr, "error: user with ID %d not found\n", userId)
			return
		}
		var user models.User
		if !driver.ReadModel(r.AppConfig.Master.FL, &user, userPos) {
			fmt.Fprintf(os.Stderr, "error: unable to calculate amount of orders\n")
			return
		}
		amount = numberOfSubrecords(r.AppConfig.Slave.FL, user.FirstOrder)

		log.Println(fmt.Sprintf("User with ID %d has %d orders", userId, amount))
		return
	}
}
func (r *Repository) CalcM(_ *cobra.Command, _ []string) {
	log.Println("Total amount of users: ", numberOfRecords(r.AppConfig.Master.Ind))
}
