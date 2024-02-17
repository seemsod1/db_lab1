package driver

import (
	"github.com/seemsod1/db_lab1/internal/models"
	"log"
	"os"
)

func FindLastNode(file *os.File, recordPos int64, model interface{}) (int64, int64) {
	var prevPos int64 = -1
	var lastPos = recordPos

	for lastPos != -1 {
		if !ReadModel(file, model, lastPos) {
			return lastPos, prevPos
		}

		switch model := model.(type) {
		case *models.Order:
			prevPos = lastPos
			lastPos = model.Next
		case *models.SHeader:
			prevPos = lastPos
			lastPos = model.Next
		default:
			log.Println("Unsupported model type")
			return -1, -1
		}

		if lastPos == -1 {
			break
		}
	}

	return lastPos, prevPos
}
