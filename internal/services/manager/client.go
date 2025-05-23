package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"log"
	"net/http"
	"orc/domain/entities"
	"strings"
	"time"
)

func (m *Manager) SelectWorker(task entities.Task) (*entities.Node, error) {
	candidates := m.Scheduler.SelectCandidateNodes(task, m.WorkerNodes)
	if candidates == nil {
		return nil, fmt.Errorf("no worker nodes found for task %s", task.ID)
	}

	scores := m.Scheduler.Score(task, candidates)
	selectedNode := m.Scheduler.Pick(scores, candidates)

	return selectedNode, nil
}

func (m *Manager) GetTasks() []*entities.Task {
	var tasks []*entities.Task
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
		time.Sleep(15 * time.Second)
	}
}

func (m *Manager) ProcessTasks() {
	for {
		log.Println("Processing any tasks in the queue")
		m.SendWork()
		time.Sleep(10 * time.Second)
	}
}

func (m *Manager) AddTask(taskEvent entities.TaskEvent) {
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
		var tasks []*entities.Task
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

			m.TaskDb[task.ID] = task
		}
	}
}

func (m *Manager) SendWork() {
	if m.Pending.Len() > 0 {
		event := m.Pending.Dequeue()
		taskEvent := event.(entities.TaskEvent)
		task := taskEvent.Task
		log.Printf("Pulled %v off pending queue\n", task)

		task.State = entities.TaskScheduled
		worker, err := m.SelectWorker(task)
		if err != nil {
			log.Printf("Error selecting worker for task %s: %v\n", task.ID, err)
		}
		taskWorker, ok := m.TaskWorkerMap[task.ID]
		if ok {
			persistedTask := m.TaskDb[task.ID]
			if taskEvent.State == entities.TaskCompleted && persistedTask.State.ValidateTransition(taskEvent.State) {
				m.stopTask(taskWorker, taskEvent.Task.ID.String())
				return
			}

		}

		m.TaskDb[task.ID] = &task

		data, err := json.Marshal(taskEvent)
		if err != nil {
			log.Printf("Unable to marshal task object: %v\n%v", task, err)
		}

		url := fmt.Sprintf("http://%s/tasks", worker.Name)
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

		taskEvent = entities.TaskEvent{}
		err = d.Decode(&taskEvent)
		if err != nil {
			fmt.Printf("Error decoding response: %s\n", err.Error())
			return
		}
		log.Printf("Received task event: %v\n", taskEvent.ID)

		m.EventDb[taskEvent.ID] = &taskEvent
		m.WorkerTaskMap[worker.Name] = append(m.WorkerTaskMap[worker.Name], task.ID)
		m.TaskWorkerMap[task.ID] = worker.Name
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

func (m *Manager) checkTaskHealth(task entities.Task) error {
	w := m.TaskWorkerMap[task.ID]
	hostPort, err := getHostPort(task.HostPorts)
	if err != nil {
		return fmt.Errorf("task %s has no exposed ports: %v", task.ID, err)
	}

	worker := strings.Split(w, ":")
	if len(worker) != 2 {
		return fmt.Errorf("invalid worker address format: %s", w)
	}

	url := fmt.Sprintf("http://%s:%s%s", worker[0], hostPort, task.HealthCheck)

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
		if task.State == entities.TaskRunning && task.RestartCount < 3 {
			err := m.checkTaskHealth(*task)
			if err != nil {
				if task.RestartCount < 3 {
					m.restartTask(task)
				}
			}
		} else if task.State == entities.TaskFailed && task.RestartCount < 3 {
			m.restartTask(task)
		}
	}
}

func (m *Manager) restartTask(task *entities.Task) {
	worker := m.TaskWorkerMap[task.ID]
	task.State = entities.TaskScheduled
	task.RestartCount++
	m.TaskDb[task.ID] = task

	taskEvent := entities.TaskEvent{
		ID:          uuid.New(),
		State:       entities.TaskRunning,
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

	newTask := entities.Task{}
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

func (m *Manager) stopTask(worker string, taskID string) {
	client := &http.Client{}
	url := fmt.Sprintf("http://%s/tasks/%s", worker, taskID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Printf("error creating request to delete task %s: %v\n", taskID, err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error deleting task %s: %v\n", taskID, err)
		return
	}
	if resp.StatusCode != http.StatusNoContent {
		log.Printf("error sending request: %v\n", err)
		return
	}
	log.Printf("task %s has been scheduled to be stopped\n", taskID)
}
