package main

import (
	"fmt"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"orc/domain/entities"
	"time"
)

func main() {
	task := entities.Task{
		ID:            uuid.New(),
		Name:          "Task-1",
		State:         entities.TaskPending,
		Image:         "Image-1",
		Memory:        1024,
		Disk:          1,
		ExposedPorts:  nil,
		PortBindings:  nil,
		RestartPolicy: "",
		StartsAt:      nil,
		FinishedAt:    nil,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	taskEvent := entities.TaskEvent{
		ID:          uuid.New(),
		State:       entities.TaskPending,
		RequestedAt: time.Now(),
		Task:        task,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	fmt.Printf("task: %v\n", task)
	fmt.Printf("task event: %v\n", taskEvent)

	worker := entities.Worker{
		Name:      "worker-1",
		Queue:     *queue.New(),
		Db:        make(map[uuid.UUID]*entities.Task),
		TaskCount: 0,
	}
	fmt.Printf("worker: %v\n", worker)

	worker.CollectStats()
	worker.RunTask()
	worker.StartTask()
	worker.StopTask()

	manager := entities.Manager{
		Pending:       *queue.New(),
		TaskDb:        make(map[string][]*entities.Task),
		EventDb:       make(map[string][]*entities.TaskEvent),
		Workers:       []string{worker.Name},
		WorkerTaskMap: nil,
		TaskWorkerMap: nil,
	}
	fmt.Printf("manager: %v\n", manager)

	manager.SelectWorker()
	manager.UpdateTasks()
	manager.SendWork()

	node := entities.Node{
		Name:            "Node-1",
		IP:              "192.168.1.1",
		Cores:           4,
		Memory:          1024,
		MemoryAllocated: 0,
		Disk:            25,
		DiskAllocated:   0,
		Role:            "worker",
		TaskCount:       0,
	}

	fmt.Printf("node: %v\n", node)
}
