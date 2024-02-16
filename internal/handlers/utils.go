package handlers

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/seemsod1/db_lab1/internal/helpers"
	"github.com/seemsod1/db_lab1/internal/models"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
)

func OrderDelete(r *Repository, readPos int64, prev int64) {
	if prev == 0 {
		prev = -1
	}
	if !helpers.DeleteOrderNode(r.AppConfig.SlaveFL, readPos, prev) {
		log.Fatal("Unable to delete order")
	}
	if !helpers.AddGarbageNode(r.AppConfig.SlaveFL, r.AppConfig.GarbageSlave, readPos, models.Order{Deleted: true}, r.AppConfig.SlavePos) {
		log.Fatal("Unable to delete order")
	}
	log.Print("Order deleted")
}
func OrderInsert(r *Repository, order models.Order, index models.IndexTable, prev ...int64) {
	var pos int64

	if r.AppConfig.GarbageSlave.Prev == -1 {
		pos = r.AppConfig.GarbageSlave.Next
	} else {
		pos = r.AppConfig.GarbageSlave.Pos
	}
	if len(prev) != 0 {
		if !helpers.AddNode(order, r.AppConfig.SlaveFL, pos, prev[0]) {
			log.Fatal("Unable to insert order")
		}
	} else {
		if index.Pos != -1 {
			index.Pos = pos
			r.AppConfig.SlaveInd = append(r.AppConfig.SlaveInd, index)
		}
		if !helpers.AddNode(order, r.AppConfig.SlaveFL, pos) {
			log.Fatal("Unable to insert order")
		}
	}

	if r.AppConfig.GarbageSlave.Prev == -1 {
		r.AppConfig.SlavePos += helpers.SlaveSize
	}
	r.AppConfig.GarbageSlave = helpers.DeleteGarbageNode(r.AppConfig.SlaveFL, r.AppConfig.GarbageSlave, r.AppConfig.GarbageSlave.Prev, r.AppConfig.SlavePos)
	if r.AppConfig.GarbageSlave == nil {
		log.Fatal("Unable to insert order")
	}
	log.Print("Order inserted")
}

func UserDelete(r *Repository, readPos int64) {
	if !helpers.AddGarbageNode(r.AppConfig.MasterFL, r.AppConfig.GarbageMaster, readPos, models.User{Deleted: true}, r.AppConfig.MasterPos) {
		log.Fatal("Unable to delete user")
	}
	log.Print("User deleted")
}

func UserInsert(r *Repository, user models.User, index models.IndexTable) {
	var pos int64
	if r.AppConfig.GarbageMaster.Prev == -1 {
		pos = r.AppConfig.GarbageMaster.Next
	} else {
		pos = r.AppConfig.GarbageMaster.Pos
		index.Pos = pos
	}
	if !helpers.WriteModel(r.AppConfig.MasterFL, &user, pos) {
		log.Fatal("Unable to insert user")
	}

	r.AppConfig.MasterInd = append(r.AppConfig.MasterInd, index)
	if r.AppConfig.GarbageMaster.Prev == -1 {
		r.AppConfig.MasterPos += helpers.MasterSize
	}
	r.AppConfig.GarbageMaster = helpers.DeleteGarbageNode(r.AppConfig.MasterFL, r.AppConfig.GarbageMaster, r.AppConfig.GarbageMaster.Prev, r.AppConfig.MasterPos)
	if r.AppConfig.GarbageMaster == nil {
		log.Fatal("Unable to insert user")
	}
	log.Print("User inserted")
}

func printMasterFile(file *os.File) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Name", "Mail", "Age", "Deleted"})

	var user models.User
	var garbage models.SHeader
	readPos := int64(0)
	for helpers.ReadModel(file, &user, readPos) {
		if user.Deleted {
			//read garbage
			if !helpers.ReadModel(file, &garbage, readPos) {
				log.Fatal("Failed to print master file")
			}
			t.AppendRow([]interface{}{-111, garbage.Prev, garbage.Pos, garbage.Next, true, "Garbage"})
			readPos += helpers.MasterSize
		} else {
			t.AppendRow([]interface{}{user.ID, helpers.ByteArrayToString(user.Name[:]), helpers.ByteArrayToString(user.Mail[:]), user.Age, user.Deleted, "User"})
			readPos += helpers.MasterSize
		}

	}

	t.Render()
}
func printSlaveFile(file *os.File) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"User_id", "Rent_id", "Price", "Country", "Deleted", "Next"})
	var order models.Order
	var garbage models.SHeader
	readPos := int64(0)
	for helpers.ReadModel(file, &order, readPos) {
		if order.Deleted {
			//read garbage
			if !helpers.ReadModel(file, &garbage, readPos) {
				log.Fatal("Failed to print slave file")
			}
			t.AppendRow([]interface{}{garbage.Prev, garbage.Pos, garbage.Next, "", true, "", "Garbage"})
			readPos += helpers.SlaveSize
		} else {
			t.AppendRow([]interface{}{order.UserId, order.RentId, order.Price, helpers.ByteArrayToString(order.Country[:]), order.Deleted, order.Next, "Order"})
			readPos += helpers.SlaveSize
		}

	}

	t.Render()
}

func printMasterRecord(file *os.File, pos int64, queries []string, all bool) {
	var readPos int64
	if all {
		readPos = helpers.MasterSize
	} else {
		readPos = pos
	}

	headers := []string{"id"}
	if len(queries) != 0 {
		for _, query := range queries {
			fieldName := strings.ToUpper(query)

			switch fieldName {
			case "ID":
			case "NAME":
				headers = append(headers, "name")
			case "MAIL":
				headers = append(headers, "mail")
			case "AGE":
				headers = append(headers, "age")
			case "DELETED":
				headers = append(headers, "deleted")
			default:
				fmt.Fprintf(os.Stderr, "field '%s' was not found\n", query)
			}
		}
	} else {
		headers = append(headers, "name", "mail", "age")
	}

	if len(headers) == 1 && !slices.Contains(queries, "id") {
		fmt.Fprintln(os.Stderr, "nothing to show")
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	headerRow := make(table.Row, len(headers))
	for i, header := range headers {
		headerRow[i] = header
	}
	t.AppendHeader(headerRow)

	for {
		var model models.User
		if !helpers.ReadModel(file, &model, readPos) {
			break
		}
		if model.Deleted {
			continue
		}

		var row []interface{}
		for _, header := range headers {
			switch strings.ToUpper(header) {
			case "ID":
				row = append(row, model.ID)
			case "NAME":
				row = append(row, helpers.ByteArrayToString(model.Name[:]))
			case "MAIL":
				row = append(row, helpers.ByteArrayToString(model.Mail[:]))
			case "AGE":
				row = append(row, model.Age)
			case "DELETED":
				row = append(row, strconv.FormatBool(model.Deleted))
			}
		}

		t.AppendRow(row)
		if !all {
			break
		}
		readPos += helpers.MasterSize
	}
	t.Render()
}
func printSlaveRecord(file *os.File, index models.IndexTable, rentId uint32, queries []string, allUsers bool, allOrders bool) {
	var readPos int64
	empty := true
	headers := []string{"user_id", "rent_id"}
	if len(queries) != 0 {
		for _, query := range queries {
			fieldName := strings.ToUpper(query)

			switch fieldName {
			case "USER_ID":
			case "RENT_ID":
			case "PRICE":
				headers = append(headers, "price")
			case "COUNTRY":
				headers = append(headers, "country")
			case "DELETED":
				headers = append(headers, "deleted")
			case "NEXT":
				headers = append(headers, "next")
			default:
				fmt.Fprintf(os.Stderr, "field '%s' was not found\n", query)
			}
		}
	} else {
		headers = append(headers, "price", "country")
	}

	if len(headers) == 2 && (!slices.Contains(queries, "user_id") || !slices.Contains(queries, "rent_id")) {
		fmt.Fprintln(os.Stderr, "nothing to show")
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	headerRow := make(table.Row, len(headers))
	for i, header := range headers {
		headerRow[i] = header
	}
	t.AppendHeader(headerRow)
	var model models.Order
	for helpers.ReadModel(file, &model, readPos) {
		if model.Deleted {
			readPos += helpers.SlaveSize
			continue
		}
		if !allUsers || !allOrders {
			if allUsers {
				if model.RentId != rentId {
					readPos += helpers.SlaveSize
					continue
				}
			}
			if allOrders {
				if model.UserId != index.UserId {
					readPos += helpers.SlaveSize
					continue
				}
			}
		}
		if !allUsers && !allOrders {
			if model.UserId != index.UserId || model.RentId != rentId {
				readPos += helpers.SlaveSize
				continue
			}
		}

		var row []interface{}
		for _, header := range headers {
			switch strings.ToUpper(header) {
			case "USER_ID":
				row = append(row, model.UserId)
			case "RENT_ID":
				row = append(row, model.RentId)
			case "PRICE":
				row = append(row, model.Price)
			case "COUNTRY":
				row = append(row, helpers.ByteArrayToString(model.Country[:]))
			case "DELETED":
				row = append(row, model.Deleted)
			case "NEXT":
				row = append(row, model.Next)
			}
		}
		t.AppendRow(row)
		empty = false
		readPos += helpers.SlaveSize
	}
	if !empty {
		t.Render()
	} else {
		fmt.Fprintln(os.Stderr, "No records found")
	}
}
