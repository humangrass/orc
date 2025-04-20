package scheduler

import "orc/domain/entities"

type Scheduler interface {
	SelectCandidateNodes(task entities.Task, nodes []*entities.Node) []*entities.Node
	Score(task entities.Task, nodes []*entities.Node) map[string]float64
	Pick(scores map[string]float64, candidates []*entities.Node) *entities.Node
}
