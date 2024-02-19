package utils

import (
	"errors"
	"fmt"
	"github.com/seemsod1/db_lab1/internal/driver"
	myErr "github.com/seemsod1/db_lab1/internal/error"
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

func GetAddressById(id uint32, ind []driver.IndexTable) int64 {
	for _, v := range ind {
		if v.Id == id {
			return v.Pos
		}
	}
	return -1
}
func GetIdByAddress(pos int64, ind []driver.IndexTable) uint32 {
	for _, v := range ind {
		if v.Pos == pos {
			return v.Id
		}
	}
	return 0
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
func FindPrevNode(file *os.File, headPos int64, recordPos int64, model interface{}) int64 {
	var prev int64
	for headPos != -1 {
		if !driver.ReadModel(file, model, headPos) {
			return -1
		}
		switch model := model.(type) {
		case *models.Order:
			if model.Next == recordPos {
				return headPos
			}
			prev = headPos
			headPos = model.Next
		case *models.SHeader:
			if model.Next == recordPos {
				return headPos
			}
			prev = headPos
			headPos = model.Next
		default:
			log.Println("Unsupported model type")
			return -1
		}
	}
	return prev

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
func SortIndicesById(indices []driver.IndexTable) []driver.IndexTable {
	sort.Slice(indices, func(i, j int) bool { return indices[i].Id < indices[j].Id })
	return indices
}
func SortIndicesByPos(indices []driver.IndexTable) []driver.IndexTable {
	sort.Slice(indices, func(i, j int) bool { return indices[i].Pos < indices[j].Pos })
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
func CloseFile(fileConfig *driver.FileConfig, isMaster bool) error {
	if len(fileConfig.Ind) == 0 {
		if err := fileConfig.FL.Truncate(0); err != nil {
			return &myErr.Error{Err: myErr.FailedToOptimize}
		}
		if isMaster {
			if !driver.WriteModel(fileConfig.FL, models.User{Deleted: true}, 0) {
				return &myErr.Error{Err: myErr.FailedToOptimize}
			}
		} else {
			if !driver.WriteModel(fileConfig.FL, models.Order{Deleted: true}, 0) {
				return &myErr.Error{Err: myErr.FailedToOptimize}
			}
		}
		garbageNode := &models.SHeader{Prev: -1, Pos: 0, Next: -1}
		if !driver.WriteModel(fileConfig.FL, garbageNode, 0) {
			return &myErr.Error{Err: myErr.FailedToOptimize}
		}
		return nil
	}
	err := optimizeFile(fileConfig, isMaster)
	if errors.Is(err, &myErr.Error{Err: myErr.AlreadyOptimized}) {
		if err = fileConfig.FL.Close(); err != nil {
			return &myErr.Error{Err: myErr.FailedToOptimize}
		}
		return &myErr.Error{Err: myErr.AlreadyOptimized}
	} else if errors.Is(err, &myErr.Error{Err: myErr.FailedToOptimize}) {
		if err = fileConfig.FL.Close(); err != nil {
			return &myErr.Error{Err: myErr.FailedToOptimize}
		}
		return &myErr.Error{Err: myErr.FailedToOptimize}
	}
	if err = fileConfig.FL.Close(); err != nil {
		return &myErr.Error{Err: myErr.FailedToOptimize}
	}
	return nil

}

func RemoveById(id uint32, indices []driver.IndexTable) []driver.IndexTable {
	for i, v := range indices {
		if v.Id == id {
			indices = append(indices[:i], indices[i+1:]...)
			break
		}
	}
	return indices
}
func ChangeMasterFirstOrder(masterFile *os.File, userPos int64, firstOrder int64) {
	var user models.User
	if !driver.ReadModel(masterFile, &user, userPos) {
		log.Fatal("Unable to change first order")
	}
	user.FirstOrder = firstOrder
	if !driver.WriteModel(masterFile, &user, userPos) {
		log.Fatal("Unable to change first order")
	}
}
func createGarbageIndexTable(file *os.File, pos int64) []driver.IndexTable {
	var indices []driver.IndexTable
	i := uint32(0)
	var garbage models.SHeader
	readPos := pos
	for readPos != -1 {
		if !driver.ReadModel(file, &garbage, readPos) {
			return nil
		}
		indices = append(indices, driver.IndexTable{Id: i, Pos: readPos})
		readPos = garbage.Next
		i++
	}
	return indices
}

func reorderGarbage(file *os.File, indices []driver.IndexTable, model any) ([]driver.IndexTable, error) {
	var garbage models.SHeader
	var ind []driver.IndexTable
	readPos := int64(0)
	if !driver.ReadModel(file, &garbage, readPos) {
		return nil, fmt.Errorf("Error: read failed")
	}
	ind = append(ind, indices[0])
	indices = append(indices[:0], indices[1:]...)
	for _, i := range indices {
		readPos = i.Pos
		if !AddGarbageNode(file, &garbage, readPos, model) {
			return nil, fmt.Errorf("Error: add garbage node failed")
		}
		ind = append(ind, driver.IndexTable{Id: i.Id, Pos: readPos})
	}
	return ind, nil
}
func NumberOfSubrecords(flFile *os.File, firstSlaveAddress int64) int {
	count := 0
	nextAddress := firstSlaveAddress

	for nextAddress != -1 {
		var slave models.Order
		if !driver.ReadModel(flFile, &slave, nextAddress) {
			fmt.Printf("error reading slave model\n")
			break
		}
		nextAddress = slave.Next
		count++
	}

	return count
}

func deleteRecords(slice []driver.IndexTable, lastUserPos int64) ([]driver.IndexTable, bool) {
	size := len(slice)
	slice = SortIndicesByPos(slice)

	for i := len(slice) - 1; i >= 0; i-- {
		if slice[i].Pos > lastUserPos {
			slice = append(slice[:i], slice[i+1:]...)
		} else {
			break
		}
	}

	return slice, size != len(slice)
}

func optimizeFile(fileConfig *driver.FileConfig, isMaster bool) error {
	gabList := createGarbageIndexTable(fileConfig.FL, 0)
	if gabList == nil || len(gabList) < 2 {
		return &myErr.Error{Err: myErr.AlreadyOptimized}
	}
	gabList = SortIndicesByPos(gabList)
	tmpInd := SortIndicesByPos(fileConfig.Ind)
	trSize := int64(0)
	var err error
	if isMaster {
		//reorder garbage
		gabList, err = reorderGarbage(fileConfig.FL, gabList, models.User{Deleted: true})
		if err != nil {
			return &myErr.Error{Err: myErr.FailedToOptimize}
		}
		trSize = driver.UserSize
	} else {
		gabList, err = reorderGarbage(fileConfig.FL, gabList, models.Order{Deleted: true})
		if err != nil {
			return &myErr.Error{Err: myErr.FailedToOptimize}
		}
		trSize = driver.OrderSize
	}
	//delete records which are after last user
	gabList, cutted := deleteRecords(gabList, tmpInd[len(tmpInd)-1].Pos)
	if cutted {
		var garbage models.SHeader
		if !driver.ReadModel(fileConfig.FL, &garbage, gabList[len(gabList)-1].Pos) {
			return &myErr.Error{Err: myErr.FailedToOptimize}
		}
		garbage.Next = -1
		if !driver.WriteModel(fileConfig.FL, garbage, gabList[len(gabList)-1].Pos) {
			return &myErr.Error{Err: myErr.FailedToOptimize}
		}
		if err := fileConfig.FL.Truncate(tmpInd[len(tmpInd)-1].Pos + trSize); err != nil {
			return &myErr.Error{Err: myErr.FailedToOptimize}
		}

	}
	return nil
}
