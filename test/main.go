package main

import (
	"log"
	"tick/client"
	tickv1 "tick/gen"
	"time"
)

type FibannaciWorkflowInput struct {
	N int64
}

type FibannaciWorkflowOutput struct {
	Result int64
}

type AddInput struct {
	A int64
	B int64
}

type AddOutput struct {
	Result int64
}

func TestWorkflow(t *tickv1.Task, inp FibannaciWorkflowInput) (*FibannaciWorkflowOutput, error) {
	var s1 int64 = 1
	var s2 int64 = 1
	var i int64

	for i = 0; i < inp.N; i++ {
		var n AddOutput
		err := client.ExecTask(t, AddTask, AddInput{A: s1, B: s2}, &n)
		if err != nil {
			return nil, err
		}

		s1 = s2
		s2 = n.Result
	}

	return &FibannaciWorkflowOutput{Result: int64(s2)}, nil
}

func AddTask(t *tickv1.Task, ai AddInput) (*AddOutput, error) {
	time.Sleep(2 * time.Second)
	return &AddOutput{Result: ai.A + ai.B}, nil
}

func TestWorkflowV2(t *tickv1.Task, fwi FibannaciWorkflowInput) (*FibannaciWorkflowOutput, error) {

	var tasks []error
	results := make([]AddOutput, fwi.N)
	for i := 0; i < int(fwi.N); i++ {
		tasks = append(tasks, client.ExecTask(t, AddTask, AddInput{A: int64(i), B: int64(i)}, &results[i]))
	}

	err := client.Parallel(tasks...)
	if err != nil {
		return nil, err
	}

	max := 0
	for _, d := range results {
		max += int(d.Result)
	}

	return &FibannaciWorkflowOutput{Result: int64(max)}, nil
}

func main() {
	client.RegisterTaskHandler(TestWorkflow)
	client.RegisterTaskHandler(AddTask)
	client.RegisterTaskHandler(TestWorkflowV2)

	if err := client.Run("http://localhost:8080", "testqueue"); err != nil {
		log.Fatalf("error starting tick client: %s", err.Error())
	}
}
