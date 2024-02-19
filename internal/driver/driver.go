package driver

import (
	"bytes"
	"encoding/binary"
	"github.com/seemsod1/db_lab1/internal/models"
	"io"
	"os"
)

type IndexTable struct {
	Pos int64
	Id  uint32
}

type FileConfig struct {
	FL          *os.File
	Pos         int64
	Ind         []IndexTable
	GarbageNode *models.SHeader
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

const FragmentationThreshold = 0.3

var IndexSize = int64(binary.Size(IndexTable{}))
var UserSize = int64(binary.Size(models.User{}))
var OrderSize = int64(binary.Size(models.Order{}))

const MinAge = 17
const MaxAge = 120

const MaxPrice = 1000000

func ReadModel(file *os.File, model any, position int64) bool {
	file.Seek(position, io.SeekStart)
	err := binary.Read(file, binary.BigEndian, model)
	if err != nil {
		return false
	}
	file.Sync()
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
