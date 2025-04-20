package manager

import (
	"fmt"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"orc/domain/core/scheduler"
	"orc/domain/entities"
)

type Manager struct {
	Pending       queue.Queue
	TaskDb        map[uuid.UUID]*entities.Task
	EventDb       map[uuid.UUID]*entities.TaskEvent
	Workers       []string
	WorkerTaskMap map[string][]uuid.UUID
	TaskWorkerMap map[uuid.UUID]string
	LastWorker    int

	WorkerNodes []*entities.Node
	Scheduler   scheduler.Scheduler
}

func NewManager(workers []string, schedulerType string) *Manager {
	taskDb := make(map[uuid.UUID]*entities.Task)
	eventDb := make(map[uuid.UUID]*entities.TaskEvent)
	workerTaskMap := make(map[string][]uuid.UUID)
	taskWorkerMap := make(map[uuid.UUID]string)
	for worker := range workers {
		workerTaskMap[workers[worker]] = []uuid.UUID{}
	}

	var nodes []*entities.Node
	for worker := range workers {
		workerTaskMap[workers[worker]] = []uuid.UUID{}

		nAPI := fmt.Sprintf("http://%v", workers[worker])
		n := entities.NewNode(workers[worker], nAPI, "worker")
		nodes = append(nodes, n)
	}

	var s scheduler.Scheduler
	switch schedulerType {
	case "roundrobin":
		s = &scheduler.RoundRobin{Name: "roundrobin"}
	default:
		s = &scheduler.RoundRobin{Name: "roundrobin"}
	}

	return &Manager{
		Pending:       *queue.New(),
		TaskDb:        taskDb,
		EventDb:       eventDb,
		Workers:       workers,
		WorkerTaskMap: workerTaskMap,
		TaskWorkerMap: taskWorkerMap,
		LastWorker:    0,
		WorkerNodes:   nodes,
		Scheduler:     s,
	}
}
