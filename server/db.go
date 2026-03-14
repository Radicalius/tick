package main

import (
	"database/sql/driver"
	"encoding/json"
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type JSONMap map[string]string

func (j JSONMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), j)
}

type TaskStatusDAL int

const (
	TASK_STATUS_SUCCESS TaskStatusDAL = 0
	TASK_STATUS_FAILURE               = 1
	TASK_STATUS_PENDING               = 2
	TASK_STATUS_TIMEOUT               = 3
)

type TaskDAL struct {
	gorm.Model

	Name       string
	Parameters JSONMap `gorm:"type:text"`
	Result     string
	Status     TaskStatusDAL `gorm:"default:2"`

	Queue      string
	ReservedAt *time.Time
	SubtaskOf  *int64
}

var db *gorm.DB

func InitDatabase() error {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	var err error
	db, err = gorm.Open(sqlite.Open("tick.db"), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return err
	}

	db.AutoMigrate(&TaskDAL{})

	return nil
}
