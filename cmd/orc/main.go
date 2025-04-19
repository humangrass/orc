package main

import (
	"fmt"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/pkg/errors" // to avoid errors from docker lib
	"log"
	"orc/domain/entities"
	mUseCase "orc/internal/usecases/manager"
	wUseCase "orc/internal/usecases/worker"
	"os"
	"strconv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	whost := os.Getenv("ORC_WORKER_HOST")
	wport, err := strconv.Atoi(os.Getenv("ORC_WORKER_PORT"))
	if err != nil {
		log.Fatal(err)
	}

	mhost := os.Getenv("ORC_MANAGER_HOST")
	mport, err := strconv.Atoi(os.Getenv("ORC_MANAGER_PORT"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Starting Orc worker-1 at %s:%d\n", whost, wport)
	fmt.Printf("Starting Orc worker-2 at %s:%d\n", whost, wport+1)
	fmt.Printf("Starting Orc worker-3 at %s:%d\n", whost, wport+2)
	fmt.Printf("Starting Orc manager at %s:%d\n", mhost, mport)

	worker1 := entities.Worker{
		Name:      "test-worker-1",
		Queue:     *queue.New(),
		Db:        make(map[uuid.UUID]*entities.Task),
		TaskCount: 0,
	}
	worker2 := entities.Worker{
		Name:      "test-worker-2",
		Queue:     *queue.New(),
		Db:        make(map[uuid.UUID]*entities.Task),
		TaskCount: 0,
	}
	worker3 := entities.Worker{
		Name:      "test-worker-3",
		Queue:     *queue.New(),
		Db:        make(map[uuid.UUID]*entities.Task),
		TaskCount: 0,
	}

	workerApi1 := wUseCase.API{
		Address: whost,
		Port:    wport,
		Worker:  &worker1,
		Router:  nil,
	}
	workerApi2 := wUseCase.API{
		Address: whost,
		Port:    wport + 1,
		Worker:  &worker2,
		Router:  nil,
	}
	workerApi3 := wUseCase.API{
		Address: whost,
		Port:    wport + 2,
		Worker:  &worker3,
		Router:  nil,
	}

	go worker1.RunTasks()
	go worker1.UpdateTasks()
	go func() {
		err = workerApi1.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	go worker2.RunTasks()
	go worker2.UpdateTasks()
	go func() {
		err = workerApi2.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	go worker3.RunTasks()
	go worker3.UpdateTasks()
	go func() {
		err = workerApi3.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	workers := []string{
		fmt.Sprintf("%s:%d", whost, wport),
		fmt.Sprintf("%s:%d", whost, wport+1),
		fmt.Sprintf("%s:%d", whost, wport+2),
	}

	manager := entities.NewManager(workers, "roundrobin")
	managerApi := mUseCase.API{
		Address: mhost,
		Port:    mport,
		Manager: manager,
		Router:  nil,
	}

	go manager.ProcessTasks()
	go manager.UpdateTasks()
	go manager.DoHealthChecks()
	err = managerApi.Start()
	if err != nil {
		log.Fatal(err)
	}
}
