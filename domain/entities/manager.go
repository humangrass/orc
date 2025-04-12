package entities

import (
	"fmt"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

type Manager struct {
	Pending       queue.Queue
	TaskDb        map[string][]*Task
	EventDb       map[string][]*TaskEvent
	Workers       []string
	WorkerTaskMap map[string][]uuid.UUID
	TaskWorkerMap map[uuid.UUID]string
}

func (m *Manager) SelectWorker() {
	fmt.Println("Select Worker...")
}

func (m *Manager) UpdateTasks() {
	fmt.Println("Update Tasks...")
}

func (m *Manager) SendWork() {
	fmt.Println("Sending work to workers")
}
