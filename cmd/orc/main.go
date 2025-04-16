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

	fmt.Printf("Starting Orc worker at %s:%d\n", whost, wport)
	fmt.Printf("Starting Orc manager at %s:%d\n", mhost, mport)

	worker := entities.Worker{
		Name:      "test-worker",
		Queue:     *queue.New(),
		Db:        make(map[uuid.UUID]*entities.Task),
		TaskCount: 0,
	}
	workers := []string{fmt.Sprintf("%s:%d", whost, wport)}
	manager := entities.NewManager(workers)

	workerApi := wUseCase.API{
		Address: whost,
		Port:    wport,
		Worker:  &worker,
		Router:  nil,
	}
	managerApi := mUseCase.API{
		Address: mhost,
		Port:    mport,
		Manager: manager,
		Router:  nil,
	}

	go worker.RunTasks()
	go worker.CollectStats()
	go worker.UpdateTasks()
	go func() {
		err = workerApi.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	go manager.ProcessTasks()
	go manager.UpdateTasks()
	go manager.DoHealthChecks()
	err = managerApi.Start()
	if err != nil {
		log.Fatal(err)
	}
}
