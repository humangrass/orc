package entities

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"log"
	"net/http"
)

type Manager struct {
	Pending       queue.Queue
	TaskDb        map[uuid.UUID]*Task
	EventDb       map[uuid.UUID]*TaskEvent
	Workers       []string
	WorkerTaskMap map[string][]uuid.UUID
	TaskWorkerMap map[uuid.UUID]string
	LastWorker    int
}

func (m *Manager) SelectWorker() string {
	var newWorker int
	if m.LastWorker+1 < len(m.Workers) {
		newWorker = m.LastWorker + 1
		m.LastWorker++
	} else {
		newWorker = 0
		m.LastWorker = 0
	}

	return m.Workers[newWorker]
}

func (m *Manager) AddTask(taskEvent TaskEvent) {
	m.Pending.Enqueue(taskEvent)
}

func (m *Manager) UpdateTasks() {
	for _, worker := range m.Workers {
		log.Printf("Checking woker %v for task updates", worker)
		url := fmt.Sprintf("http://%s/tasks", worker)
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Error connecting to %v: %v\n", worker, err)
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("Error sending request: %v\n", err)
		}

		d := json.NewDecoder(resp.Body)
		var tasks []*Task
		err = d.Decode(&tasks)
		if err != nil {
			log.Printf("Error unmarshalling tasks: %s\n", err.Error())
		}

		for _, task := range tasks {
			log.Printf("Attempting to update task %v\n", task.ID)

			_, ok := m.TaskDb[task.ID]
			if !ok {
				log.Printf("Task with ID %s not found\n", task.ID)
				return
			}

			if m.TaskDb[task.ID].State != task.State {
				m.TaskDb[task.ID].State = task.State
			}

			m.TaskDb[task.ID].StartsAt = task.StartsAt
			m.TaskDb[task.ID].FinishedAt = task.FinishedAt
			m.TaskDb[task.ID].ContainerID = task.ContainerID
		}
	}
}

func (m *Manager) SendWork() {
	if m.Pending.Len() > 0 {
		worker := m.SelectWorker()
		event := m.Pending.Dequeue()
		taskEvent := event.(TaskEvent)
		task := taskEvent.Task
		log.Printf("Pulled %v off pending queue\n", task)

		m.EventDb[taskEvent.ID] = &taskEvent
		m.WorkerTaskMap[worker] = append(m.WorkerTaskMap[worker], taskEvent.Task.ID)
		m.TaskWorkerMap[task.ID] = worker

		task.State = TaskScheduled
		m.TaskDb[task.ID] = &task

		data, err := json.Marshal(taskEvent)
		if err != nil {
			log.Printf("Unable to marshal task object: %v\n%v", task, err)
		}

		url := fmt.Sprintf("http://%s/tasks", worker)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
		if err != nil {
			log.Printf("Error connecting to %v: %v\n", worker, err)
			m.Pending.Enqueue(taskEvent)
		}
		d := json.NewDecoder(resp.Body)
		if resp.StatusCode != http.StatusCreated {
			e := ErrResponse{}
			err = d.Decode(&e)
			if err != nil {
				fmt.Printf("Error decoding response: %s\n", err.Error())
				return
			}
			log.Printf("Response error (%d): %s\n", e.HTTPStatusCode, e.Message)
			return
		}

		task = Task{}
		err = d.Decode(&task)
		if err != nil {
			fmt.Printf("Error decoding response: %s\n", err.Error())
			return
		}
		log.Printf("Received task: %v\n", task)
	} else {
		log.Println("No work in the queue")
	}
}

func NewManager(workers []string) *Manager {
	taskDb := make(map[uuid.UUID]*Task)
	eventDb := make(map[uuid.UUID]*TaskEvent)
	workerTaskMap := make(map[string][]uuid.UUID)
	taskWorkerMap := make(map[uuid.UUID]string)
	for worker := range workers {
		workerTaskMap[workers[worker]] = []uuid.UUID{}
	}

	return &Manager{
		Pending:       *queue.New(),
		TaskDb:        taskDb,
		EventDb:       eventDb,
		Workers:       workers,
		WorkerTaskMap: workerTaskMap,
		TaskWorkerMap: taskWorkerMap,
		LastWorker:    0,
	}
}
