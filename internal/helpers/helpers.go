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

func OpenMasterFile(filename string, withTrunc bool) (*os.File, int64) {

	FL, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	if withTrunc {
		WriteModel(FL, models.User{}, 0)
		return FL, MasterSize
	}
	return FL, 0
}
func OpenSlaveFile(filename string, withTrunc bool) (*os.File, int64) {

	FL, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	if withTrunc {
		WriteModel(FL, models.Order{}, 0)
		return FL, int64(unsafe.Sizeof(models.Order{}))
	}
	return FL, 0
}

func PrintSlave(slave models.Order) {
	fmt.Println(fmt.Sprintf("UserId: %d\nRentId: %d\nPrice: %d\n", slave.UserId, slave.RentId, slave.Price))
}
func PrintMaster(master models.User) {
	fmt.Println(fmt.Sprintf("Id: %d\nName: %s\nMail: %s\nAge: %d\n", master.ID, master.Name, master.Mail, master.Age))
}
func GetPosition(id uint32, ind []models.IndexTable) int64 {
	for _, v := range ind {
		if v.UserId == id {
			return v.Pos
		}
	}
	return -1
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
