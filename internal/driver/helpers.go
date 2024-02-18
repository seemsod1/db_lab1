package driver

import (
	"encoding/binary"
	"github.com/seemsod1/db_lab1/internal/models"
	"os"
)

func FindLastNode(file *os.File, recordPos int64, model interface{}) int64 {
	var tmp int64

	for {
		if !ReadModel(file, model, recordPos) {
			return -1
		}
		switch modelTmp := model.(type) {
		case *models.SHeader:
			tmp = modelTmp.Next
		case *models.Order:
			tmp = modelTmp.Next
		default:
			return -1
		}
		if tmp == -1 {
			break
		}
		recordPos = tmp
	}
	return recordPos
}

func FindFilePos(file *os.File, model interface{}) int64 {
	var pos int64 = 0
	for {
		if !ReadModel(file, model, pos) {
			break
		}
		pos += int64(binary.Size(model))
	}
	return pos
}
