package utils

import (
	"fmt"
	"github.com/seemsod1/db_lab1/internal/driver"
	"github.com/seemsod1/db_lab1/internal/models"
	"log"
	"os"
	"sort"
	"strings"
)

func SetNextPtr(file *os.File, recordPos int64, nextRecordPos int64) bool {
	var tmp models.Order

	if !driver.ReadModel(file, &tmp, recordPos) {
		log.Println("Error: Read failed")
		return false
	}

	tmp.Next = nextRecordPos

	if !driver.WriteModel(file, tmp, recordPos) {
		log.Println("Error: Write failed")
		return false
	}
	return true
}

func AddNode(record models.Order, file *os.File, pos int64, prevRecordPos ...int64) bool {
	var prev int64
	if len(prevRecordPos) == 0 {
		prev = -1
	} else {
		prev = prevRecordPos[0]
	}
	if !driver.WriteModel(file, record, pos) {
		log.Println("Error: write failed")
		return false
	}

	if prev == -1 {
		return true
	}

	return SetNextPtr(file, prev, pos)
}

func GetAddressByIndex(id uint32, ind []driver.IndexTable) (int64, int) {
	for i, v := range ind {
		if v.UserId == id {
			return v.Pos, i
		}
	}
	return -1, -1
}

func DeleteOrderNode(file *os.File, recordPos int64, prevRecordPos int64) bool {
	var tmp models.Order
	if !driver.ReadModel(file, &tmp, recordPos) {
		log.Println("Error: read failed")
		return false
	}

	if prevRecordPos == -1 {
		tmp.Deleted = true
		if !driver.WriteModel(file, tmp, recordPos) {
			log.Println("Error: write failed")
			return false
		}

		return true
	}

	var prev models.Order
	if !driver.ReadModel(file, &prev, prevRecordPos) {
		log.Println("Error: read failed")
		return false
	}

	prev.Next = tmp.Next
	if !driver.WriteModel(file, prev, prevRecordPos) {
		log.Println("Error: write failed")
		return false
	}
	tmp.Deleted = true
	if !driver.WriteModel(file, tmp, recordPos) {
		log.Println("Error: write failed")
		return false
	}
	return true
}

func AddGarbageNode(file *os.File, garbageSlave *models.SHeader, readPos int64, data any) bool {
	garbageSlave.Next = readPos
	if !driver.WriteModel(file, garbageSlave, garbageSlave.Pos) {
		log.Println("Error: write failed")
		return false
	}
	if !driver.WriteModel(file, data, readPos) {
		log.Println("Error: write failed")
		return false
	}
	garbageSlave.Prev = garbageSlave.Pos
	garbageSlave.Pos = readPos
	garbageSlave.Next = -1
	if !driver.WriteModel(file, garbageSlave, readPos) {
		log.Println("Error: write failed")
		return false
	}
	return true
}

func DeleteGarbageNode(fileConfig *driver.FileConfig) *models.SHeader {

	if fileConfig.GarbageNode.Prev == -1 {
		fileConfig.GarbageNode.Next = -1
		if !driver.WriteModel(fileConfig.FL, fileConfig.GarbageNode, fileConfig.GarbageNode.Pos) {
			log.Println("Error: write failed")
			return nil
		}
		return fileConfig.GarbageNode
	}

	var prev models.SHeader
	if !driver.ReadModel(fileConfig.FL, &prev, fileConfig.GarbageNode.Prev) {
		fmt.Println("Error: read failed")
		return nil
	}

	prev.Next = fileConfig.GarbageNode.Next
	if !driver.WriteModel(fileConfig.FL, prev, fileConfig.GarbageNode.Prev) {
		fmt.Println("Error: write failed")
		return nil
	}
	return &prev
}

func ByteArrayToString(bytes []byte) string {
	return strings.TrimRight(string(bytes), "\x00")
}

func NumberOfRecords(indices []driver.IndexTable) int {
	return len(indices)
}
func SortIndices(indices []driver.IndexTable) []driver.IndexTable {
	sort.Slice(indices, func(i, j int) bool { return indices[i].UserId < indices[j].UserId })
	return indices
}

func CalculateAmountOfNodes(file *os.File, headPos int64, model interface{}) int {
	var amount int

	for headPos != -1 {
		if !driver.ReadModel(file, model, headPos) {
			return -1
		}
		amount++
		switch model := model.(type) {
		case *models.Order:
			headPos = model.Next
		case *models.SHeader:
			headPos = model.Next
		default:
			log.Println("Unsupported model type")
			return -1
		}
	}
	return amount
}

func WriteIndices(filename string, indices []driver.IndexTable) {
	FL, err := os.OpenFile(filename+".ind", os.O_RDWR|os.O_CREATE, 0666)
	// File was opened
	if err = FL.Truncate(0); err != nil {
		log.Fatal(err)
	}
	writePos := int64(0)
	for _, v := range indices {
		if !driver.WriteModel(FL, v, writePos) {
			log.Fatal(err)
		}
		writePos += driver.IndexSize
	}
	log.Println("Indices written")
	FL.Close()
}
func CloseFile(fileConfig *driver.FileConfig, isMaster bool) bool {
	if len(fileConfig.Ind) == 0 {
		fileConfig.FL.Truncate(0)
		if isMaster {
			if !driver.WriteModel(fileConfig.FL, models.User{Deleted: true}, 0) {
				return false
			}
		} else {
			if !driver.WriteModel(fileConfig.FL, models.Order{Deleted: true}, 0) {
				return false
			}
		}
		garbageNode := &models.SHeader{Prev: -1, Pos: 0, Next: -1}
		if !driver.WriteModel(fileConfig.FL, garbageNode, 0) {
			return false
		}

	}
	//optimizing file
	if isMaster {
		//optimizing master file
	} else {

	}

	fileConfig.FL.Close()
	return true
}
