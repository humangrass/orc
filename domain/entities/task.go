package entities

import (
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"time"
)

type TaskState int

const (
	TaskPending TaskState = iota
	TaskScheduled
	TaskRunning
	TaskCompleted
	TaskFailed = iota - 5
)

var taskStateTransitionMap = map[TaskState][]TaskState{
	TaskPending:   {TaskScheduled},
	TaskScheduled: {TaskScheduled, TaskRunning, TaskFailed},
	TaskRunning:   {TaskRunning, TaskCompleted, TaskFailed},
	TaskCompleted: {},
	TaskFailed:    {},
}

func (s *TaskState) ValidateTransition(destination TaskState) bool {
	allowed, exists := taskStateTransitionMap[*s]
	if !exists {
		return false
	}
	for _, state := range allowed {
		if state == destination {
			return true
		}
	}
	return false
}

type Task struct {
	ID            uuid.UUID
	ContainerID   string
	Name          string
	State         TaskState
	Image         string
	Memory        int
	Disk          int
	ExposedPorts  nat.PortSet
	PortBindings  map[string]string
	RestartPolicy string
	StartsAt      *time.Time
	FinishedAt    *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type TaskEvent struct {
	ID          uuid.UUID
	State       TaskState
	RequestedAt time.Time
	Task        Task
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
