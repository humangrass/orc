package main

import (
	"fmt"
	"github.com/docker/docker/client"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	_ "github.com/pkg/errors" // to avoid errors from docker lib
	"orc/domain/entities"
	"os"
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

	fmt.Println("create a test container")
	dockerTask, createResult := createContainer()
	if createResult.Error != nil {
		fmt.Printf("createResult: %v\n", createResult)
		os.Exit(1)
	}
	time.Sleep(time.Second * 5)
	fmt.Printf("stopping container %s\n\n", createResult.ContainerID)
	_ = stopContainer(dockerTask, createResult.ContainerID)
}

func createContainer() (*entities.Docker, *entities.DockerResult) {
	config := entities.OrcConfig{
		Name:  "test-container-1",
		Image: "postgres:13",
		Env: []string{
			"POSTGRES_USER=cube",
			"POSTGRES_PASSWORD=secret",
		},
	}

	dc, _ := client.NewClientWithOpts(client.FromEnv)
	d := entities.Docker{
		Client: dc,
		Config: config,
	}

	result := d.Run()
	if result.Error != nil {
		fmt.Printf("%v\n\n", result.Error)
		return nil, nil
	}

	fmt.Printf("Container %s is running with config %v\n", result.ContainerID, config)
	return &d, &result
}

func stopContainer(d *entities.Docker, id string) *entities.DockerResult {
	result := d.Stop(id)
	if result.Error != nil {
		fmt.Printf("Error stopping container %s: %v\n", id, result.Error)
		return &entities.DockerResult{Error: result.Error}
	}

	fmt.Printf("Container %s is stopped and removed\n", result.ContainerID)
	return &result
}
