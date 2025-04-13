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

	fmt.Println("Starting Orc worker")
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
	err = api.Start()
	if err != nil {
		log.Fatal(err)
	}
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
