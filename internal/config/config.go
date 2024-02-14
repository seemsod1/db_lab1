package config

import (
	"github.com/seemsod1/db_lab1/internal/models"
	"os"
)

type AppConfig struct {
	MasterFL  *os.File
	SlaveFL   *os.File
	MasterPos int64
	SlavePos  int64
	MasterInd []models.IndexTable
	SlaveInd  []models.IndexTable

	GarbageMaster *models.SHeader
	GarbageSlave  *models.SHeader
}
