package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	tickv1 "tick/gen"
	"tick/gen/tickv1connect"
	"time"
)

var client tickv1connect.TickClient
var handlerRegistry map[string]func(*tickv1.Task) (string, error) = make(map[string]func(*tickv1.Task) (string, error))

func RegisterTaskHandler[I any, O any](name string, handler func(*tickv1.Task, I) (*O, error)) {
	handlerRegistry[name] = func(t *tickv1.Task) (string, error) {
		var input I
		if err := json.Unmarshal([]byte(t.Parameters), &input); err != nil {
			return "", err
		}

		out, err := handler(t, input)
		if err != nil {
			return "", err
		}

		data, err := json.Marshal(out)
		if err != nil {
			return "", err
		}

		return string(data), nil
	}
}

func Parallel(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

func ExecTask[I any, O any](parent *tickv1.Task, name string, parameter I, output *O) error {
	paramStr, err := json.Marshal(parameter)
	if err != nil {
		return err
	}

	res, err := client.GetTask(context.Background(), &tickv1.GetTaskRequest{
		TaskName:   name,
		Parameters: string(paramStr),
	})
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			_, err := client.QueueTask(context.Background(), &tickv1.QueueTaskRequest{
				QueueName:  globalQueueName,
				TaskName:   name,
				Parameters: string(paramStr),
				ParentId:   parent.TaskId,
			})
			if err != nil {
				return err
			}

			return fmt.Errorf("task execution deferred")
		}

		return err
	}

	if res.Task.Status == tickv1.TaskStatus_SUCCESS {
		if err := json.Unmarshal([]byte(res.Task.Result), output); err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("task execution deferred")
}

func execTask(task *tickv1.Task) error {
	var result string
	var status = tickv1.TaskStatus_FAILURE

	handler, exists := handlerRegistry[task.TaskName]
	if exists {
		var err error
		result, err = handler(task)
		if err != nil {
			if strings.Contains(err.Error(), "task execution deferred") {
				status = tickv1.TaskStatus_PENDING
			} else {
				result = err.Error()
			}
		} else {
			status = tickv1.TaskStatus_SUCCESS
		}
	} else {
		result = "task not registered"
	}

	_, err := client.ResolveTask(context.Background(), &tickv1.ResolveTaskRequest{
		TaskId: task.TaskId,
		Result: result,
		Status: status,
	})

	return err
}

var globalQueueName string

func Run(addr string, queueName string) error {
	globalQueueName = queueName
	client = tickv1connect.NewTickClient(http.DefaultClient, addr)

	for {
		res, err := client.PollTaskQueue(context.Background(), &tickv1.PollTaskQueueRequest{
			QueueName: queueName,
		})
		if err != nil {
			return err
		}

		if res.Task == nil {
			time.Sleep(5 * time.Second)
			continue
		}

		err = execTask(res.Task)
		if err != nil {
			return err
		}
	}
}
