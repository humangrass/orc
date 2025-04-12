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

type Task struct {
	ID            uuid.UUID
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
