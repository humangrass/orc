package worker

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"log"
	"net/http"
	"orc/domain/entities"
)

type API struct {
	Address string
	Port    int
	Worker  *entities.Worker
	Router  *chi.Mux
}

type ErrResponse struct {
	Message        string `json:"message"`
	HTTPStatusCode int    `json:"-"`
}

func (a *API) initRouter() {
	a.Router = chi.NewRouter()
	a.Router.Route("/tasks", func(r chi.Router) {
		r.Post("/", a.StartTaskHandler)
		r.Get("/", a.GetTaskHandler)
		r.Route("/{taskID}", func(r chi.Router) {
			r.Delete("/", a.StopTaskHandler)
		})
	})
	a.Router.Route("/stats", func(r chi.Router) {
		r.Get("/", a.GetStatsHandler)
	})
}

func (a *API) Start() error {
	a.initRouter()
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", a.Address, a.Port), a.Router)
	return err
}

func (a *API) GetStatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(a.Worker.Stats)
	if err != nil {
		log.Println(err)
		return
	}
}

func (a *API) StartTaskHandler(w http.ResponseWriter, r *http.Request) {
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	taskEvent := entities.TaskEvent{}
	err := d.Decode(&taskEvent)
	if err != nil {
		msg := fmt.Sprintf("Error unmarshalling body: %v\n", err)
		log.Println(msg)
		w.WriteHeader(http.StatusBadRequest)
		e := ErrResponse{
			HTTPStatusCode: http.StatusBadRequest,
			Message:        msg,
		}
		err := json.NewEncoder(w).Encode(e)
		if err != nil {
			log.Println(err)
			return
		}
		return
	}
	a.Worker.AddTask(taskEvent.Task)
	log.Printf("Task added: %v\n", taskEvent.Task.ID)
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(taskEvent)
	if err != nil {
		log.Println(err)
		return
	}
}

func (a *API) GetTaskHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(a.Worker.GetTasks())
	if err != nil {
		log.Println(err)
		return
	}
}

func (a *API) StopTaskHandler(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	if taskID == "" {
		log.Printf("Invalid task ID: %v\n", taskID)
		w.WriteHeader(http.StatusBadRequest)
	}

	tID, err := uuid.Parse(taskID)
	if err != nil {
		log.Printf("Invalid task ID: %v\n", taskID)
		w.WriteHeader(http.StatusBadRequest)
	}

	_, ok := a.Worker.Db[tID]
	if !ok {
		log.Printf("Task not found: %v\n", tID)
		w.WriteHeader(http.StatusNotFound)
	}

	taskToStop := a.Worker.Db[tID]
	taskCopy := *taskToStop
	taskCopy.State = entities.TaskCompleted
	a.Worker.AddTask(taskCopy)

	log.Printf("Added task %v to stop container %v\n", taskToStop.ID, taskToStop.ContainerID)
	w.WriteHeader(http.StatusNoContent)
}
