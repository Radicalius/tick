package main

import (
	"context"
	"fmt"
	tickv1 "tick/gen"
	"tick/gen/tickv1connect"
)

type TickServiceServer struct {
	tickv1connect.UnimplementedTickHandler
}

func (t *TickServiceServer) QueueTask(context context.Context, request *tickv1.QueueTaskRequest) (*tickv1.QueueTaskResponse, error) {
	err := db.Create(&TaskDAL{
		Name:       request.Task.TaskName,
		Parameters: request.Task.Parameters,
		Queue:      request.QueueName,
	}).Error
	if err != nil {
		return nil, err
	}

	return &tickv1.QueueTaskResponse{}, nil
}

func (t *TickServiceServer) ResolveTask(context context.Context, request *tickv1.ResolveTaskRequest) (*tickv1.ResolveTaskResponse, error) {
	task := TaskDAL{}
	if err := db.First(&task, request.TaskId).Error; err != nil {
		return nil, err
	}

	if task.Status != TaskStatusDAL(TASK_STATUS_PENDING) {
		return nil, fmt.Errorf("task already resolved")
	}

	var translatedStatus TaskStatusDAL
	switch request.Status {
	case tickv1.TaskStatus_FAILURE:
		translatedStatus = TaskStatusDAL(TASK_STATUS_FAILURE)
	case tickv1.TaskStatus_SUCCESS:
		translatedStatus = TaskStatusDAL(TASK_STATUS_SUCCESS)
	default:
		return nil, fmt.Errorf("invalid task status: %d", request.Status)
	}

	err := db.Model(&task).Updates(map[string]interface{}{
		"result": request.Result,
		"status": translatedStatus,
	}).Error

	if err != nil {
		return nil, err
	}

	return &tickv1.ResolveTaskResponse{}, nil
}

func (t *TickServiceServer) GetTask(context context.Context, request *tickv1.GetTaskRequest) (*tickv1.GetTaskResponse, error) {
	task := TaskDAL{}
	if err := db.First(&task, "name = ? and parameters = ?", request.TaskName, JSONMap(request.Parameters)).Error; err != nil {
		return nil, err
	}

	var status tickv1.TaskStatus
	switch task.Status {
	case TaskStatusDAL(TASK_STATUS_FAILURE):
		status = tickv1.TaskStatus_FAILURE
	case TaskStatusDAL(TASK_STATUS_TIMEOUT):
		status = tickv1.TaskStatus_FAILURE
	case TaskStatusDAL(TASK_STATUS_SUCCESS):
		status = tickv1.TaskStatus_SUCCESS
	case TaskStatusDAL(TASK_STATUS_PENDING):
		status = tickv1.TaskStatus_PENDING
	}

	return &tickv1.GetTaskResponse{
		Task: &tickv1.Task{
			TaskName:   task.Name,
			Parameters: request.Parameters,
			Result:     task.Result,
			Status:     status,
		},
	}, nil
}

func (t *TickServiceServer) PollTaskQueue(context.Context, *tickv1.PollTaskQueueRequest) (*tickv1.PollTaskQueueResponse, error) {
	return &tickv1.PollTaskQueueResponse{}, nil
}
