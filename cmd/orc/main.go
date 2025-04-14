package main

import (
	"fmt"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/pkg/errors" // to avoid errors from docker lib
	"log"
	"orc/domain/entities"
	"os"
	"strconv"
	"time"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	host := os.Getenv("ORC_HOST")
	port, err := strconv.Atoi(os.Getenv("ORC_PORT"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Starting Orc")
	fmt.Printf("host: %s, port: %d\n", host, port)

	worker := entities.Worker{
		Name:      "test-worker",
		Queue:     *queue.New(),
		Db:        make(map[uuid.UUID]*entities.Task),
		TaskCount: 0,
	}
	api := entities.API{
		Address: host,
		Port:    port,
		Worker:  &worker,
		Router:  nil,
	}

	go runTasks(&worker)
	go worker.CollectStats()
	go func() {
		err = api.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	time.Sleep(10 * time.Second)
	workers := []string{fmt.Sprintf("%s:%d", host, port)}
	manager := entities.NewManager(workers)
	for i := 0; i < 3; i++ {
		task := entities.Task{
			ID:          uuid.New(),
			ContainerID: "",
			Name:        fmt.Sprintf("test-container-%d", i),
			State:       entities.TaskScheduled,
			Image:       "strm/helloworld-http",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		taskEvent := entities.TaskEvent{
			ID:          uuid.New(),
			State:       entities.TaskRunning,
			RequestedAt: time.Time{},
			Task:        task,
		}

		manager.AddTask(taskEvent)
		manager.SendWork()
	}

	//go func() {
	//	for {
	//		fmt.Printf("[Manager] Updating task from %d workers\n", len(manager.Workers))
	//		time.Sleep(15 * time.Second)
	//	}
	//}()
	//for {
	//	for _, t := range manager.TaskDb {
	//		fmt.Printf("[Manager] Task: id: %s, state: %d\n", t.ID, t.State)
	//		time.Sleep(15 * time.Second)
	//	}
	//}
}

func runTasks(worker *entities.Worker) {
	for {
		if worker.Queue.Len() != 0 {
			result := worker.RunTask()
			if result.Error != nil {
				log.Printf("Error running task: %v\n", result.Error)
			}
		} else {
			log.Println("No tasks left")
		}
		log.Println("Sleeping for 10 seconds.")
		time.Sleep(10 * time.Second)
	}
}
