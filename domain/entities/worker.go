package entities

import (
	"fmt"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

type Worker struct {
	Name      string
	Queue     queue.Queue
	Db        map[uuid.UUID]*Task
	TaskCount int
}

func (w *Worker) CollectStats() {
	fmt.Println("Collecting stats")
}

func (w *Worker) RunTask() {
	fmt.Println("Starting or stopping a task")
}

func (w *Worker) StartTask() {
	fmt.Println("Starting a task")
}

func (w *Worker) StopTask() {
	fmt.Println("Stopping a task")
}
