package entities

import (
	"fmt"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"log"
	"time"
)

type Worker struct {
	Name      string
	Queue     queue.Queue
	Db        map[uuid.UUID]*Task
	TaskCount int
	Stats     *Stats
}

func (w *Worker) CollectStats() {
	for {
		log.Println("Collecting stats...")
		w.Stats = GetStats()
		w.Stats.TaskCount = w.TaskCount
		time.Sleep(15 * time.Second)
	}
}

func (w *Worker) RunTask() DockerResult {
	t := w.Queue.Dequeue()
	if t == nil {
		log.Println("Queue is empty")
		return DockerResult{Error: nil}
	}

	taskQueued := t.(Task)
	taskPersisted := w.Db[taskQueued.ID]
	if taskPersisted == nil {
		taskPersisted = &taskQueued
		w.Db[taskQueued.ID] = taskPersisted
	}

	var result DockerResult
	if taskPersisted.State.ValidateTransition(taskQueued.State) {
		switch taskQueued.State {
		case TaskScheduled:
			result = w.StartTask(taskQueued)
		case TaskCompleted:
			result = w.StopTask(taskQueued)
		default:
			result.Error = errors.New("unreachable code")
		}
	} else {
		err := fmt.Errorf("invalid transition from %v to %v", taskQueued.State, taskQueued.State)
		result.Error = err
	}

	return result
}

func (w *Worker) AddTask(task Task) {
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	w.Queue.Enqueue(task)
}

func (w *Worker) RunTasks() {
	for {
		if w.Queue.Len() != 0 {
			result := w.RunTask()
			if result.Error != nil {
				log.Printf("Error running task: %v", result.Error)
			}
		} else {
			log.Println("No tasks to process currently")
		}
		time.Sleep(10 * time.Second)
	}
}

func (w *Worker) StartTask(t Task) DockerResult {
	now := time.Now()
	t.StartsAt = &now
	t.UpdatedAt = now
	config := NewOrcConfig(&t)
	d, err := NewDocker(config)
	if err != nil {
		return DockerResult{Error: err}
	}

	result := d.Run()
	if result.Error != nil {
		log.Printf("Err running task: %v: %v\n", t.ID, result.Error)
		t.State = TaskFailed
		w.Db[t.ID] = &t
		return result
	}

	t.ContainerID = result.ContainerID
	t.State = TaskRunning
	w.Db[t.ID] = &t

	return result
}

func (w *Worker) StopTask(t Task) DockerResult {
	config := NewOrcConfig(&t)
	d, err := NewDocker(config)
	if err != nil {
		return DockerResult{Error: err}
	}

	result := d.Stop(t.ContainerID)
	if result.Error != nil {
		log.Printf("Error stopping container %v: %v\n", t.ContainerID, result.Error)
	}
	now := time.Now()

	t.FinishedAt = &now
	t.UpdatedAt = now
	t.State = TaskCompleted
	w.Db[t.ID] = &t
	log.Printf("Stopped and removed container %v for task %v\n", t.ContainerID, t.ID)

	return result
}

func (w *Worker) GetTasks() []Task {
	tasks := make([]Task, 0, len(w.Db))
	for _, task := range w.Db {
		tasks = append(tasks, *task)
	}
	return tasks
}

func (w *Worker) InspectTask(task Task) DockerInspectResponse {
	config := NewOrcConfig(&task)
	docker, err := NewDocker(config)
	if err != nil {
		return DockerInspectResponse{Error: err}
	}

	return docker.Inspect(task.ContainerID)
}

func (w *Worker) UpdateTasks() {
	for {
		log.Println("Checking status of tasks")
		w.updateTasks()
		time.Sleep(15 * time.Second)
	}
}

func (w *Worker) updateTasks() {
	for id, task := range w.Db {
		if task.State == TaskRunning {
			resp := w.InspectTask(*task)
			if resp.Error != nil {
				fmt.Printf("ERROR: %v\n", resp.Error)
			}

			if resp.Container == nil {
				log.Printf("No container for running task %v\n", id)
				w.Db[id].State = TaskFailed
			}
			if resp.Container.State.Status == "exited" {
				log.Printf("Container %v is exited\n", id)
				w.Db[id].State = TaskFailed
			}

			w.Db[id].HostPorts = resp.Container.NetworkSettings.Ports
		}
	}
}
