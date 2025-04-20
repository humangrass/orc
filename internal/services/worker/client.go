package worker

import (
	"fmt"
	"github.com/pkg/errors"
	"log"
	"orc/domain/entities"
	"orc/internal/infrastructure/docker"
	"orc/pkg/xstats"
	"time"
)

func (w *Worker) CollectStats() {
	for {
		log.Println("Collecting stats...")
		w.Stats = xstats.GetStats()
		w.Stats.TaskCount = w.TaskCount
		time.Sleep(15 * time.Second)
	}
}

func (w *Worker) RunTask() docker.Result {
	t := w.Queue.Dequeue()
	if t == nil {
		log.Println("Queue is empty")
		return docker.Result{Error: nil}
	}

	taskQueued := t.(entities.Task)
	taskPersisted := w.Db[taskQueued.ID]
	if taskPersisted == nil {
		taskPersisted = &taskQueued
		w.Db[taskQueued.ID] = taskPersisted
	}

	var result docker.Result
	if taskPersisted.State.ValidateTransition(taskQueued.State) {
		switch taskQueued.State {
		case entities.TaskScheduled:
			result = w.StartTask(taskQueued)
		case entities.TaskCompleted:
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

func (w *Worker) AddTask(task entities.Task) {
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

func (w *Worker) StartTask(t entities.Task) docker.Result {
	now := time.Now()
	t.StartsAt = &now
	config := entities.NewOrcConfig(&t)
	d, err := docker.NewDocker(config)
	if err != nil {
		return docker.Result{Error: err}
	}

	result := d.Run()
	if result.Error != nil {
		log.Printf("Err running task: %v: %v\n", t.ID, result.Error)
		t.State = entities.TaskFailed
		w.Db[t.ID] = &t
		return result
	}

	t.ContainerID = result.ContainerID
	t.State = entities.TaskRunning
	w.Db[t.ID] = &t

	return result
}

func (w *Worker) StopTask(t entities.Task) docker.Result {
	config := entities.NewOrcConfig(&t)
	d, err := docker.NewDocker(config)
	if err != nil {
		return docker.Result{Error: err}
	}

	result := d.Stop(t.ContainerID)
	if result.Error != nil {
		log.Printf("Error stopping container %v: %v\n", t.ContainerID, result.Error)
	}
	now := time.Now()

	t.FinishedAt = &now
	t.State = entities.TaskCompleted
	w.Db[t.ID] = &t
	log.Printf("Stopped and removed container %v for task %v\n", t.ContainerID, t.ID)

	return result
}

func (w *Worker) GetTasks() []entities.Task {
	tasks := make([]entities.Task, 0, len(w.Db))
	for _, task := range w.Db {
		tasks = append(tasks, *task)
	}
	return tasks
}

func (w *Worker) InspectTask(task entities.Task) docker.InspectResponse {
	config := entities.NewOrcConfig(&task)
	d, err := docker.NewDocker(config)
	if err != nil {
		return docker.InspectResponse{Error: err}
	}

	return d.Inspect(task.ContainerID)
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
		if task.State == entities.TaskRunning {
			resp := w.InspectTask(*task)
			if resp.Error != nil {
				fmt.Printf("ERROR: %v\n", resp.Error)
			}

			if resp.Container == nil {
				log.Printf("No container for running task %v\n", id)
				w.Db[id].State = entities.TaskFailed
			}
			if resp.Container.State.Status == "exited" {
				log.Printf("Container %v is exited\n", id)
				w.Db[id].State = entities.TaskFailed
			}

			w.Db[id].HostPorts = resp.Container.NetworkSettings.Ports
		}
	}
}
