package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type TimeSeriesResp struct {
	Timestamp int64  `json:"timestamp"`
	Name      string `json:"name"`
	Queue     string `json:"queue"`
	Status    string `json:"status"`
	SubtaskOf *int64 `json:"suntaskOf"`
	Count     int64  `json:"count"`
}

type TaskResp struct {
	Timestamp  int64  `json:"timestamp"`
	Body       string `json:"body"`
	Level      string `json:"level"`
	Name       string `json:"name"`
	Parameters string `json:"parameters"`
	Result     string `json:"result"`
	Status     string `json:"status"`
	Queue      string `json:"queue"`
	SubtaskOf  *int64 `json:"subtaskOf"`
}

func taskRespFromTaskDal(dal *TaskDAL) *TaskResp {
	strStatus := "UNKNOWN"
	level := "info"
	switch dal.Status {
	case TaskStatusDAL(TASK_STATUS_FAILURE):
		strStatus = "FAILURE"
		level = "error"
	case TaskStatusDAL(TASK_STATUS_TIMEOUT):
		strStatus = "TIMEOUT"
		level = "warning"
	case TaskStatusDAL(TASK_STATUS_SUCCESS):
		strStatus = "SUCCESS"
		level = "info"
	case TaskStatusDAL(TASK_STATUS_PENDING):
		strStatus = "PENDING"
		level = "debug"
	}

	return &TaskResp{
		Timestamp:  dal.CreatedAt.UnixMilli(),
		Body:       fmt.Sprintf("[%s] %s::%s(%s) -> %s", strStatus, dal.Queue, dal.Name, dal.Parameters, dal.Result),
		Level:      level,
		Name:       dal.Name,
		Parameters: dal.Parameters,
		Result:     dal.Result,
		Status:     strStatus,
		Queue:      dal.Queue,
		SubtaskOf:  dal.SubtaskOf,
	}
}

type Filters struct {
	Start     time.Time
	End       time.Time
	Name      string
	Queue     string
	Status    string
	SubtaskOf int
	Groups    []string
}

func (f *Filters) Filter(tx *gorm.DB) *gorm.DB {
	tx = tx.Where("created_at between ? and ?", f.Start, f.End)
	if f.Name != "" {
		tx = tx.Where("name = ?", f.Name)
	}
	if f.Queue != "" {
		tx = tx.Where("queue = ?", f.Queue)
	}
	if f.Status != "" {
		tx = tx.Where("status = ?", f.Status)
	}
	if f.SubtaskOf != 0 {
		tx = tx.Where("sub_task_of = ?", f.SubtaskOf)
	}
	return tx
}

func (f *Filters) GroupBy(tx *gorm.DB) *gorm.DB {
	for _, group := range f.Groups {
		tx = tx.Group(group)
	}

	return tx
}

func (f *Filters) Select(tx *gorm.DB) *gorm.DB {
	if len(f.Groups) > 0 {
		return tx.Select("(unixepoch(created_at) - (unixepoch(created_at) % 60)) * 1000 as timestamp, COUNT(*) as count, " + strings.Join(f.Groups, ", "))
	}

	return tx.Select("(unixepoch(created_at) - (unixepoch(created_at) % 60)) * 1000 as timestamp, COUNT(*) as count")
}

func parseFilters(r *http.Request) (*Filters, error) {
	startQ := r.URL.Query().Get("start")
	start, err := strconv.ParseInt(startQ, 10, 64)
	if err != nil {
		fmt.Printf("error parsing start timestamp: %s\n", err.Error())
		return nil, err
	}

	endQ := r.URL.Query().Get("end")
	end, err := strconv.ParseInt(endQ, 10, 64)
	if err != nil {
		fmt.Printf("error parsing end timestamp: %s\n", err.Error())
		return nil, err
	}

	startTime := time.UnixMilli(start)
	endTime := time.UnixMilli(end)

	name := r.URL.Query().Get("name")
	queue := r.URL.Query().Get("queue")
	status := r.URL.Query().Get("status")

	var subtask int64
	subtaskQ := r.URL.Query().Get("subtaskOf")
	if subtaskQ != "" {
		subtask, err = strconv.ParseInt(subtaskQ, 10, 64)
		if err != nil {
			fmt.Printf("error parsing subtask: %s\n", err.Error())
			return nil, err
		}
	}

	return &Filters{
		Start:     startTime,
		End:       endTime,
		Name:      name,
		Queue:     queue,
		Status:    status,
		SubtaskOf: int(subtask),
		Groups:    r.URL.Query()["groupBy"],
	}, nil
}

func RunMetricsApi() {
	http.HandleFunc("/api/v1/history", func(w http.ResponseWriter, r *http.Request) {
		filters, err := parseFilters(r)
		if err != nil {
			w.WriteHeader(500)
			return
		}

		var tasks []TaskDAL
		err = filters.Filter(db).
			Order("created_at desc").
			Find(&tasks).Error
		if err != nil {
			fmt.Printf("error reading history logs: %s\n", err.Error())
			w.WriteHeader(500)
			return
		}

		taskResps := make([]*TaskResp, 0)
		for _, task := range tasks {
			taskResps = append(taskResps, taskRespFromTaskDal(&task))
		}

		json.NewEncoder(w).Encode(taskResps)
	})

	http.HandleFunc("/api/v1/counts", func(w http.ResponseWriter, r *http.Request) {
		filters, err := parseFilters(r)
		if err != nil {
			w.WriteHeader(500)
			return
		}

		var timeSeries []TimeSeriesResp
		err = filters.Select(filters.Filter(filters.GroupBy(db.Model(&TaskDAL{})))).
			Group("unixepoch(created_at) - (unixepoch(created_at) % 60)").
			Order("timestamp").
			Find(&timeSeries).Error
		if err != nil {
			fmt.Printf("error querying metric counts %s", err)
			w.WriteHeader(500)
			return
		}

		json.NewEncoder(w).Encode(timeSeries)
	})

	http.ListenAndServe(":8001", nil)
}
