package sceduler

type Scheduler interface {
	SelectCandidateNodes()
	Store()
	Pick()
}
