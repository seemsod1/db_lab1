package driver

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/seemsod1/db_lab1/internal/models"
	"io"
	"log"
	"os"
)

type IndexTable struct {
	Pos    int64
	UserId uint32
}

type FileConfig struct {
	FL          *os.File
	Pos         int64
	Ind         []IndexTable
	GarbageNode *models.SHeader
	Size        int
	IndSize     int
}

func NewFileConfig(file *os.File, pos int64, ind []IndexTable, garbageNode *models.SHeader) *FileConfig {
	return &FileConfig{
		FL:          file,
		Pos:         pos,
		Ind:         ind,
		GarbageNode: garbageNode,
	}
}

const MasterFilename = "user"
const SlaveFilename = "order"

var IndexSize = int64(binary.Size(IndexTable{}))
var UserSize = int64(binary.Size(models.User{}))
var OrderSize = int64(binary.Size(models.Order{}))

const MinAge = 17

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

func CreateFileConfig(filename string, isMaster bool) (*FileConfig, error) {
	FL, err := os.OpenFile(filename+".fl", os.O_RDWR, 0666)
	var headerNext int64
	var indices []IndexTable
	var garbageNode *models.SHeader
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// File doesn't exist, it was created
			FL, err = os.OpenFile(filename+".fl", os.O_RDWR|os.O_CREATE, 0666)
			if isMaster {
				if !WriteModel(FL, models.User{Deleted: true}, 0) {
					return nil, err
				}
				headerNext = UserSize
			} else {
				if !WriteModel(FL, models.Order{Deleted: true}, 0) {
					return nil, err
				}
				headerNext = OrderSize
			}
			garbageNode = &models.SHeader{Prev: -1, Pos: 0, Next: headerNext}
			if !WriteModel(FL, garbageNode, 0) {
				return nil, err
			}

			log.Println("New config created")
		} else {
			// Some other error occurred
			return nil, err
		}
	} else {
		// File was opened
		ind, err := os.OpenFile(filename+".ind", os.O_RDWR, 0666)
		if err != nil {
			log.Fatal(err)
		}

		indices, err = LoadIndices(ind)
		if err != nil {
			return nil, err
		}
		posInFile, prev := FindLastNode(FL, 0, &models.SHeader{})
		if prev == -1 {
			return nil, err
		}
		var gab models.SHeader
		if !ReadModel(FL, &gab, prev) {
			return nil, err
		}
		garbageNode = &gab
		headerNext = posInFile

		log.Println("Config loaded")
	}

	fileConfig := NewFileConfig(FL, headerNext, indices, garbageNode)
	return fileConfig, nil
}
func LoadIndices(indFile *os.File) ([]IndexTable, error) {
	readPos := int64(0)

	var indices []IndexTable
	for {
		var model IndexTable
		if !ReadModel(indFile, &model, readPos) {
			break
		}
		readPos += IndexSize
		indices = append(indices, model)
	}

	return indices, nil
}
