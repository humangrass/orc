package worker

import (
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"orc/domain/entities"
	"orc/pkg/xstats"
)

type Worker struct {
	Name      string
	Queue     queue.Queue
	Db        map[uuid.UUID]*entities.Task
	TaskCount int
	Stats     *xstats.Stats
}
