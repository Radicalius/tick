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
	client.RegisterTaskHandler[FibannaciWorkflowInput, FibannaciWorkflowOutput](
		"testworkflow",
		func(t *tickv1.Task, inp FibannaciWorkflowInput) (*FibannaciWorkflowOutput, error) {
			var s1 int64 = 1
			var s2 int64 = 1
			var i int64

			for i = 0; i < inp.N; i++ {
				n, err := client.ExecTask[AddInput, AddOutput](t, "addtask", AddInput{A: s1, B: s2})
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

	if err := client.Run("http://localhost:8080", "testqueue"); err != nil {
		log.Fatalf("error starting tick client: %s", err.Error())
	}
}
