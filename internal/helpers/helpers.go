package helpers

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/seemsod1/db_lab1/internal/models"
	"io"
	"log"
	"os"
	"unsafe"
)

const MasterSize = int64(unsafe.Sizeof(models.User{}))
const SlaveSize = int64(unsafe.Sizeof(models.Order{}))
const HeaderSize = int64(unsafe.Sizeof(models.SHeader{}))

// TODO: EOF
func ReadModel(file *os.File, model any, position int64) bool {
	file.Seek(position, io.SeekStart)
	err := binary.Read(file, binary.BigEndian, model)
	if err != nil {
		return false
	}
	return true
}

func WriteModel(file *os.File, model any, position int64) bool {
	var binBuf bytes.Buffer
	err := binary.Write(&binBuf, binary.BigEndian, model)
	if err != nil {
		return false
	}

	file.Seek(position, io.SeekStart)
	_, err = file.Write(binBuf.Bytes())

	if err != nil {
		log.Fatal(err)
	}

	file.Sync()
	return true
}

func SetNextPtr(file *os.File, recordPos int64, nextRecordPos int64) bool {
	var tmp models.Order

	if !ReadModel(file, &tmp, recordPos) {
		fmt.Println("Unable to set next ptr. Read failed")
		return false
	}

	tmp.Next = nextRecordPos

	if !WriteModel(file, tmp, recordPos) {
		fmt.Println("Unable to set next ptr. Read failed")
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
	if !WriteModel(file, record, pos) {
		return false
	}

	if prev == -1 {
		return true
	}

	return SetNextPtr(file, prev, pos)
}

func PrintNodes(file *os.File, recordPos int64) {
	var tmp models.Order

	readPos := recordPos
	fmt.Println("Linked List:")
	for readPos != -1 {
		if !ReadModel(file, &tmp, readPos) {
			fmt.Println("Unable to update next_ptr. Error: read failed")
			return
		}

		fmt.Println(fmt.Sprintf("UserId: %d\nRentId: %d\nPrice: %d\n", tmp.UserId, tmp.RentId, tmp.Price))
		readPos = tmp.Next
	}
}

func OpenSlaveFile(filename string) (*os.File, int64, *models.SHeader) {

	FL, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	WriteModel(FL, models.Order{Deleted: true}, 0)

	WriteModel(FL, models.SHeader{Prev: -1, Pos: 0, Next: SlaveSize}, 0)
	return FL, SlaveSize, &models.SHeader{Prev: -1, Pos: 0, Next: SlaveSize}
}

func OpenMasterFile(filename string) (*os.File, int64, *models.SHeader) {

	FL, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	WriteModel(FL, models.User{Deleted: true}, 0)

	WriteModel(FL, models.SHeader{Prev: -1, Pos: 0, Next: MasterSize}, 0)
	return FL, MasterSize, &models.SHeader{Prev: -1, Pos: 0, Next: MasterSize}
}

//	func PrintSlave(slave models.Order) {
//		fmt.Println(fmt.Sprintf("UserId: %d\nRentId: %d\nPrice: %d\n", slave.UserId, slave.RentId, slave.Price))
//	}
//
// TODO: FIX UT-S COUNTRY
func PrintSlave(slave models.Order, full bool) {
	if full {
		//str := string(slave.Country[:])
		fmt.Printf("| %-10d | %-10d | %-10d | %-10s | %-10t | %-7d |\n", slave.UserId, slave.RentId, slave.Price, "country", slave.Deleted, slave.Next)
	} else {
		fmt.Printf("| %-10d | %-10d | %-10d | %-10s |\n", slave.UserId, slave.RentId, slave.Price, "country")
	}
}

func PrintGarbage(garbage models.SHeader) {
	fmt.Printf("| %-10d | %-10d | %-10d | %-10s | %-10t | %-7d |\n", garbage.Prev, garbage.Pos, garbage.Next, "garbagge", true, 0)
}

func PrintMaster(master models.User, full bool) {
	if full {
		fmt.Println(fmt.Sprintf("Id: %d\nName: %s\nMail: %s\nAge: %d\n", master.ID, master.Name, master.Mail, master.Age))
	} else {
		fmt.Println(fmt.Sprintf("Id: %d\nName: %s\nMail: %s\n", master.ID, master.Name, master.Mail))
	}
}
func GetPosition(id uint32, ind []models.IndexTable) (int64, int) {
	for i, v := range ind {
		if v.UserId == id {
			return v.Pos, i
		}
	}
	return -1, -1
}

func FindLastNode(file *os.File, recordPos int64) int64 {
	var tmp models.Order

	readPos := recordPos
	for readPos != -1 {
		if !ReadModel(file, &tmp, readPos) {
			fmt.Println("Unable to update next_ptr. Error: read failed")
			return -1
		}
		if tmp.Next == -1 {
			break
		}
		readPos = tmp.Next
	}
	return readPos
}

func DeleteOrderNode(file *os.File, recordPos int64, prevRecordPos int64) {
	var tmp models.Order
	if !ReadModel(file, &tmp, recordPos) {
		fmt.Println("Unable to delete node. Error: read failed")
		return
	}

	if prevRecordPos == -1 {
		tmp.Deleted = true
		WriteModel(file, tmp, recordPos)
		return
	}

	var prev models.Order
	if !ReadModel(file, &prev, prevRecordPos) {
		fmt.Println("Unable to delete node. Error: read failed")
		return
	}

	prev.Next = tmp.Next
	WriteModel(file, prev, prevRecordPos)
	tmp.Deleted = true
	WriteModel(file, tmp, recordPos)
}

func AddSHeaderNode(record models.SHeader, file *os.File, pos int64, prevRecordPos ...int64) bool {
	var prev int64
	if len(prevRecordPos) == 0 {
		prev = -1
	} else {
		prev = prevRecordPos[0]
	}
	if !WriteModel(file, record, pos) {
		return false
	}

	if prev == -1 {
		return true
	}

	return SetNextGarbagePtr(file, prev, pos)
}

func SetNextGarbagePtr(file *os.File, recordPos int64, nextRecordPos int64) bool {
	var tmp models.SHeader

	if !ReadModel(file, &tmp, recordPos) {
		fmt.Println("Unable to set next ptr. Read failed")
		return false
	}

	tmp.Next = nextRecordPos

	if !WriteModel(file, tmp, recordPos) {
		fmt.Println("Unable to set next ptr. Read failed")
		return false
	}
	return true
}
