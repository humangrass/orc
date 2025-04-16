package entities

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"log"
	"net/http"
	"strings"
	"time"
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

func (m *Manager) GetTasks() []*Task {
	var tasks []*Task
	for _, task := range m.TaskDb {
		tasks = append(tasks, task)
	}
	return tasks
}

func (m *Manager) UpdateTasks() {
	time.Sleep(10 * time.Second)
	for {
		log.Println("Checking for task updates from workers")
		m.updateTasks()
		log.Println("Task updates completed")
		log.Println("Sleeping for 15 second")
		time.Sleep(15 * time.Second)
	}
}

func (m *Manager) ProcessTasks() {
	for {
		log.Println("Processing any tasks in the queue")
		m.SendWork()
		log.Println("Sleeping for 10 second")
		time.Sleep(10 * time.Second)
	}
}

func (m *Manager) AddTask(taskEvent TaskEvent) {
	m.Pending.Enqueue(taskEvent)
}

func (m *Manager) updateTasks() {
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

func getHostPort(ports nat.PortMap) (string, error) {
	if len(ports) == 0 {
		return "", fmt.Errorf("no ports found")
	}
	for _, bindings := range ports {
		if len(bindings) > 0 {
			return bindings[0].HostPort, nil
		}
	}

	return "", fmt.Errorf("no host ports found")
}

func (m *Manager) checkTaskHealth(task Task) error {
	w := m.TaskWorkerMap[task.ID]
	hostPort, err := getHostPort(task.HostPorts)
	if err != nil {
		return fmt.Errorf("task %s has no exposed ports: %v", task.ID, err)
	}

	worker := strings.Split(w, ":")
	if len(worker) != 2 {
		return fmt.Errorf("invalid worker address format: %s", w)
	}

	url := fmt.Sprintf("http://%s/%s%s", worker[0], hostPort, task.HealthCheck)

	log.Printf("Calling health check for task %s: %s\n", task.ID, url)
	resp, err := http.Get(url)
	if err != nil {
		msg := fmt.Sprintf("Error connecting to health check %s: %v\n", url, err)
		log.Printf(msg)
		return fmt.Errorf(msg)
	}

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("Error health chech status code for task %s\n", task.ID)
		log.Printf(msg)
		return fmt.Errorf(msg)
	}
	log.Printf("Health check for task %s\n", task.ID)

	return nil
}

func (m *Manager) doHealthCheck() {
	for _, task := range m.GetTasks() {
		if task.State == TaskRunning && task.RestartCount < 3 {
			err := m.checkTaskHealth(*task)
			if err != nil {
				if task.RestartCount < 3 {
					m.restartTask(task)
				}
			}
		} else if task.State == TaskFailed && task.RestartCount < 3 {
			m.restartTask(task)
		}
	}
}

func (m *Manager) restartTask(task *Task) {
	worker := m.TaskWorkerMap[task.ID]
	task.State = TaskScheduled
	task.RestartCount++
	m.TaskDb[task.ID] = task

	taskEvent := TaskEvent{
		ID:          uuid.New(),
		State:       TaskRunning,
		RequestedAt: time.Now(),
		Task:        *task,
	}
	data, err := json.Marshal(taskEvent)
	if err != nil {
		log.Printf("Unable to marshal task object: %v\n%v", taskEvent, err)
		return
	}

	url := fmt.Sprintf("http://%s/tasks", worker)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Error connecting to %v: %v\n", worker, err)
		m.Pending.Enqueue(taskEvent)
		return
	}

	d := json.NewDecoder(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		e := ErrResponse{}
		err := d.Decode(&e)
		if err != nil {
			fmt.Printf("Error decoding response: %s\n", err.Error())
			return
		}
		log.Printf("Response error (%d): %s\n", e.HTTPStatusCode, e.Message)
		return
	}

	newTask := Task{}
	err = d.Decode(&newTask)
	if err != nil {
		fmt.Printf("Error decoding response: %s\n", err.Error())
		return
	}
	log.Printf("Received task: %v\n", newTask)
}

func (m *Manager) DoHealthChecks() {
	for {
		log.Println("Performing health checks")
		m.doHealthCheck()
		time.Sleep(30 * time.Second)
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
