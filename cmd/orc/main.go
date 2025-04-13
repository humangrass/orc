package main

import (
	"fmt"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	_ "github.com/pkg/errors" // to avoid errors from docker lib
	"log"
	"orc/domain/entities"
	"time"
)

func main() {
	db := make(map[uuid.UUID]*entities.Task)
	worker := entities.Worker{
		Queue: *queue.New(),
		Db:    db,
	}
	task := entities.Task{
		ID:        uuid.New(),
		Name:      "test-helloworld-container-1",
		State:     entities.TaskScheduled,
		Image:     "strm/helloworld-http",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	fmt.Println("starting task")
	worker.AddTask(task)

	worker.CollectStats()
	result := worker.RunTask()
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	task.ContainerID = result.ContainerID
	fmt.Printf("task %s is running in container %s\n", task.ID, task.ContainerID)
	fmt.Println("Sleepy time!")
	time.Sleep(30 * time.Second)

	fmt.Printf("stopping task %s\n", task.ID)
	task.State = entities.TaskCompleted
	worker.AddTask(task)
	result = worker.RunTask()
	if result.Error != nil {
		log.Fatal(result.Error)
	}
}
