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
}

func (w *Worker) CollectStats() {
	fmt.Println("Collecting stats")
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
	w.Queue.Enqueue(task)
}

func (w *Worker) StartTask(t Task) DockerResult {
	now := time.Now()
	t.StartsAt = &now
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
	t.State = TaskCompleted
	w.Db[t.ID] = &t
	log.Printf("Stopped and removed container %v for task %v\n", t.ContainerID, t.ID)

	return result
}
