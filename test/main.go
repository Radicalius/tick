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

func main() {
	client.RegisterTaskHandler(
		"testworkflow",
		func(t *tickv1.Task, inp FibannaciWorkflowInput) (*FibannaciWorkflowOutput, error) {
			var s1 int64 = 1
			var s2 int64 = 1
			var i int64

			for i = 0; i < inp.N; i++ {
				var n AddOutput
				err := client.ExecTask(t, "addtask", AddInput{A: s1, B: s2}, &n)
				if err != nil {
					return nil, err
				}

				s1 = s2
				s2 = n.Result
			}

			return &FibannaciWorkflowOutput{Result: int64(s2)}, nil
		},
	)

	client.RegisterTaskHandler[AddInput, AddOutput]("addtask", func(t *tickv1.Task, ai AddInput) (*AddOutput, error) {
		time.Sleep(2 * time.Second)
		return &AddOutput{Result: ai.A + ai.B}, nil
	})

	client.RegisterTaskHandler(
		"testworkflow.v2",
		func(t *tickv1.Task, fwi FibannaciWorkflowInput) (*FibannaciWorkflowOutput, error) {

			var tasks []error
			results := make([]AddOutput, fwi.N)
			for i := 0; i < int(fwi.N); i++ {
				tasks = append(tasks, client.ExecTask(t, "addtask", AddInput{A: int64(i), B: int64(i)}, &results[i]))
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
		},
	)

	if err := client.Run("http://localhost:8080", "testqueue"); err != nil {
		log.Fatalf("error starting tick client: %s", err.Error())
	}
}
