package helpers

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/seemsod1/db_lab1/internal/models"
	"io"
	"log"
	"os"
	"strings"
	"unsafe"
)

const MasterSize = int64(unsafe.Sizeof(models.User{}))
const SlaveSize = int64(unsafe.Sizeof(models.Order{}))
const MaxAge = 18

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
		return false
	}

	file.Sync()
	return true
}

func SetNextPtr(file *os.File, recordPos int64, nextRecordPos int64) bool {
	var tmp models.Order

	if !ReadModel(file, &tmp, recordPos) {
		fmt.Println("Error: Read failed")
		return false
	}

	tmp.Next = nextRecordPos

	if !WriteModel(file, tmp, recordPos) {
		fmt.Println("Error: Write failed")
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
		fmt.Println("Error: write failed")
		return false
	}

	if prev == -1 {
		return true
	}

	return SetNextPtr(file, prev, pos)
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

	FL, err := os.OpenFile(filename+".fl", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	WriteModel(FL, models.User{Deleted: true}, 0)

	WriteModel(FL, models.SHeader{Prev: -1, Pos: 0, Next: MasterSize}, 0)
	return FL, MasterSize, &models.SHeader{Prev: -1, Pos: 0, Next: MasterSize}
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
			fmt.Println("Error: read failed")
			return -1
		}
		if tmp.Next == -1 {
			break
		}
		readPos = tmp.Next
	}
	return readPos
}

func DeleteOrderNode(file *os.File, recordPos int64, prevRecordPos int64) bool {
	var tmp models.Order
	if !ReadModel(file, &tmp, recordPos) {
		fmt.Println("Error: read failed")
		return false
	}

	if prevRecordPos == -1 {
		tmp.Deleted = true
		if !WriteModel(file, tmp, recordPos) {
			fmt.Println("Error: write failed")
			return false
		}

		return true
	}

	var prev models.Order
	if !ReadModel(file, &prev, prevRecordPos) {
		fmt.Println("Error: read failed")
		return false
	}

	prev.Next = tmp.Next
	if !WriteModel(file, prev, prevRecordPos) {
		fmt.Println("Error: write failed")
		return false
	}
	tmp.Deleted = true
	if !WriteModel(file, tmp, recordPos) {
		fmt.Println("Error: write failed")
		return false
	}
	return true
}

func AddGarbageNode(file *os.File, garbageSlave *models.SHeader, readPos int64, data any, posInFile int64) bool {
	garbageSlave.Next = readPos
	if !WriteModel(file, garbageSlave, garbageSlave.Pos) {
		fmt.Println("Error: write failed")
		return false
	}
	if !WriteModel(file, data, readPos) {
		fmt.Println("Error: write failed")
		return false
	}
	garbageSlave.Prev = garbageSlave.Pos
	garbageSlave.Pos = readPos
	garbageSlave.Next = posInFile
	if !WriteModel(file, garbageSlave, readPos) {
		fmt.Println("Error: write failed")
		return false
	}
	return true
}

func DeleteGarbageNode(file *os.File, actualNode *models.SHeader, prevRecordPos int64, posInFile int64) *models.SHeader {

	if actualNode.Prev == -1 {
		actualNode.Next = posInFile
		if !WriteModel(file, actualNode, actualNode.Pos) {
			fmt.Println("Error: write failed")
			return nil
		}
		return actualNode
	}

	var prev models.SHeader
	if !ReadModel(file, &prev, prevRecordPos) {
		fmt.Println("Error: read failed")
		return nil
	}

	prev.Next = actualNode.Next
	if !WriteModel(file, prev, prevRecordPos) {
		fmt.Println("Error: write failed")
		return nil
	}
	return &prev
}

func ByteArrayToString(bytes []byte) string {
	return strings.TrimRight(string(bytes), "\x00")
}

func NumberOfRecords(indices []models.IndexTable) int {
	return len(indices)
}
