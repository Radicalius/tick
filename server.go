package main

import (
	"context"
	tickv1 "tick/gen"
	"tick/gen/tickv1connect"
)

type TickServiceServer struct {
	tickv1connect.UnimplementedTickHandler
}

func (t *TickServiceServer) QueueTask(context.Context, *tickv1.QueueTaskRequest) (*tickv1.QueueTaskResponse, error) {
	return &tickv1.QueueTaskResponse{}, nil
}

func (t *TickServiceServer) ResolveTask(context.Context, *tickv1.ResolveTaskRequest) (*tickv1.ResolveTaskResponse, error) {
	return &tickv1.ResolveTaskResponse{}, nil
}

func (t *TickServiceServer) PollTaskQueue(context.Context, *tickv1.PollTaskQueueRequest) (*tickv1.PollTaskQueueResponse, error) {
	return &tickv1.PollTaskQueueResponse{}, nil
}
