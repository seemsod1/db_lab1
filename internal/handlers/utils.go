package handlers

import (
	"encoding/binary"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/seemsod1/db_lab1/internal/driver"
	"github.com/seemsod1/db_lab1/internal/driver/utils"
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
	if !utils.DeleteOrderNode(r.AppConfig.Slave.FL, readPos, prev) {
		log.Fatal("Unable to delete order")
	}
	if !utils.AddGarbageNode(r.AppConfig.Slave.FL, r.AppConfig.Slave.GarbageNode, readPos, models.Order{Deleted: true}, r.AppConfig.Slave.Pos) {
		log.Fatal("Unable to delete order")
	}
	log.Print("Order deleted")
}
func OrderInsert(r *Repository, order models.Order, index driver.IndexTable, prev ...int64) {
	var pos int64

	if r.AppConfig.Slave.GarbageNode.Prev == -1 {
		pos = r.AppConfig.Slave.GarbageNode.Next
	} else {
		pos = r.AppConfig.Slave.GarbageNode.Pos
	}
	if len(prev) != 0 {
		if !utils.AddNode(order, r.AppConfig.Slave.FL, pos, prev[0]) {
			log.Fatal("Unable to insert order")
		}
	} else {
		if index.Pos != -1 {
			index.Pos = pos
			r.AppConfig.Slave.Ind = append(r.AppConfig.Slave.Ind, index)
			r.AppConfig.Slave.Ind = utils.SortIndices(r.AppConfig.Slave.Ind)
		}
		if !utils.AddNode(order, r.AppConfig.Slave.FL, pos) {
			log.Fatal("Unable to insert order")
		}
	}

	if r.AppConfig.Slave.GarbageNode.Prev == -1 {
		r.AppConfig.Slave.Pos += driver.OrderSize
	}
	r.AppConfig.Slave.GarbageNode = utils.DeleteGarbageNode(r.AppConfig.Slave.FL, r.AppConfig.Slave.GarbageNode, r.AppConfig.Slave.GarbageNode.Prev, r.AppConfig.Slave.Pos)
	if r.AppConfig.Slave.GarbageNode == nil {
		log.Fatal("Unable to insert order")
	}
	log.Print("Order inserted")
}

func UserDelete(r *Repository, readPos int64) {
	if !utils.AddGarbageNode(r.AppConfig.Master.FL, r.AppConfig.Master.GarbageNode, readPos, models.User{Deleted: true}, r.AppConfig.Master.Pos) {
		log.Fatal("Unable to delete user")
	}
	log.Print("User deleted")
}
func UserInsert(r *Repository, user models.User, index driver.IndexTable) {
	var pos int64
	if r.AppConfig.Master.GarbageNode.Prev == -1 {
		pos = r.AppConfig.Master.GarbageNode.Next
	} else {
		pos = r.AppConfig.Master.GarbageNode.Pos
		index.Pos = pos
	}
	if !driver.WriteModel(r.AppConfig.Master.FL, &user, pos) {
		log.Fatal("Unable to insert user")
	}

	r.AppConfig.Master.Ind = append(r.AppConfig.Master.Ind, index)
	r.AppConfig.Master.Ind = utils.SortIndices(r.AppConfig.Master.Ind)
	if r.AppConfig.Master.GarbageNode.Prev == -1 {
		r.AppConfig.Master.Pos += driver.UserSize
	}
	r.AppConfig.Master.GarbageNode = utils.DeleteGarbageNode(r.AppConfig.Master.FL, r.AppConfig.Master.GarbageNode, r.AppConfig.Master.GarbageNode.Prev, r.AppConfig.Master.Pos)
	if r.AppConfig.Master.GarbageNode == nil {
		log.Fatal("Unable to insert user")
	}
	log.Print("User inserted")
}

func printMasterFile(file *os.File) {
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Name", "Mail", "Age", "Deleted"})

	var user models.User
	var garbage models.SHeader
	readPos := int64(0)
	for driver.ReadModel(file, &user, readPos) {
		if user.Deleted {
			//read garbage
			if !driver.ReadModel(file, &garbage, readPos) {
				log.Fatal("Failed to print master file")
			}
			t.AppendRow([]interface{}{-111, garbage.Prev, garbage.Pos, garbage.Next, true, "Garbage", driver.UserSize})
			readPos += driver.UserSize
		} else {
			t.AppendRow([]interface{}{user.ID, utils.ByteArrayToString(user.Name[:]), utils.ByteArrayToString(user.Mail[:]), user.Age, user.Deleted, "User", binary.Size(user)})
			readPos += driver.UserSize
		}

	}

	t.Render()
}
func printSlaveFile(file *os.File) {
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"User_id", "Rent_id", "Price", "Country", "Deleted", "Next"})
	var order models.Order
	var garbage models.SHeader
	readPos := int64(0)
	for driver.ReadModel(file, &order, readPos) {
		if order.Deleted {
			//read garbage
			if !driver.ReadModel(file, &garbage, readPos) {
				log.Fatal("Failed to print slave file")
			}
			t.AppendRow([]interface{}{garbage.Prev, garbage.Pos, garbage.Next, "", true, "", "Garbage", driver.OrderSize})
		} else {
			t.AppendRow([]interface{}{order.UserId, order.RentId, order.Price, utils.ByteArrayToString(order.Country[:]), order.Deleted, order.Next, "Order", binary.Size(order)})
		}
		readPos += driver.OrderSize
	}

	t.Render()
}

func printMasterRecord(file *os.File, indexTable []driver.IndexTable, pos int64, queries []string, all bool) {
	var readPos int64
	ind := 0
	if all {
		readPos = indexTable[ind].Pos
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
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	headerRow := make(table.Row, len(headers))
	for i, header := range headers {
		headerRow[i] = header
	}
	t.AppendHeader(headerRow)

	for {
		var model models.User
		if !driver.ReadModel(file, &model, readPos) {
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
				row = append(row, utils.ByteArrayToString(model.Name[:]))
			case "MAIL":
				row = append(row, utils.ByteArrayToString(model.Mail[:]))
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
		ind++
		if ind >= len(indexTable) {
			break
		}
		readPos = indexTable[ind].Pos
	}
	t.Render()
}
func printSlaveRecord(file *os.File, index driver.IndexTable, rentId uint32, queries []string, allUsers bool, allOrders bool) {
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
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	headerRow := make(table.Row, len(headers))
	for i, header := range headers {
		headerRow[i] = header
	}
	t.AppendHeader(headerRow)
	var model models.Order
	for driver.ReadModel(file, &model, readPos) {
		if model.Deleted {
			readPos += driver.OrderSize
			continue
		}
		if !allUsers || !allOrders {
			if allUsers {
				if model.RentId != rentId {
					readPos += driver.OrderSize
					continue
				}
			}
			if allOrders {
				if model.UserId != index.UserId {
					readPos += driver.OrderSize
					continue
				}
			}
		}
		if !allUsers && !allOrders {
			if model.UserId != index.UserId || model.RentId != rentId {
				readPos += driver.OrderSize
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
				row = append(row, utils.ByteArrayToString(model.Country[:]))
			case "DELETED":
				row = append(row, model.Deleted)
			case "NEXT":
				row = append(row, model.Next)
			}
		}
		t.AppendRow(row)
		empty = false
		readPos += driver.OrderSize
	}
	if !empty {
		t.Render()
	} else {
		fmt.Fprintln(os.Stderr, "No records found")
	}
}
