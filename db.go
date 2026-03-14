package main

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TaskStatus int

const (
	TASK_STATUS_SUCCESS TaskStatus = iota
	TASK_STATUS_FAILURE
	TASK_STATUS_PENDING
	TASK_STATUS_TIMEOUT
)

type TaskDAL struct {
	gorm.Model

	Name       string
	Parameters string
	Result     string
	Status     TaskStatus

	Queue      string
	ReservedAt time.Time
	SubtaskOf  *int
}

var db *gorm.DB

func InitDatabase() error {
	var err error
	db, err = gorm.Open(sqlite.Open("tick.db"), &gorm.Config{})
	if err != nil {
		return err
	}

	db.AutoMigrate(&TaskDAL{})

	return nil
}
