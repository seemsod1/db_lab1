package config

import (
	"github.com/seemsod1/db_lab1/internal/driver"
)

type AppConfig struct {
	Master *driver.FileConfig
	Slave  *driver.FileConfig
}
