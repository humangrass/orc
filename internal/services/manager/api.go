package manager

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"log"
	"net/http"
	"orc/domain/entities"
	"time"
)

type API struct {
	Address string
	Port    int
	Manager *Manager
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
}

func (a *API) Start() error {
	a.initRouter()
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", a.Address, a.Port), a.Router)
	return err
}

func (a *API) GetTasksHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(a.Manager.GetTasks())
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
	a.Manager.AddTask(taskEvent)
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
	err := json.NewEncoder(w).Encode(a.Manager.GetTasks())
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

	taskToStop, ok := a.Manager.TaskDb[tID]
	if !ok {
		log.Printf("Task not found: %v\n", tID)
		w.WriteHeader(http.StatusNotFound)
	}

	taskCopy := *taskToStop
	taskCopy.State = entities.TaskCompleted

	taskEvent := entities.TaskEvent{
		ID:          uuid.New(),
		State:       entities.TaskCompleted,
		RequestedAt: time.Now(),
		Task:        taskCopy,
	}
	a.Manager.AddTask(taskEvent)

	log.Printf("Added task %v to stop container %v\n", taskToStop.ID, taskToStop.ContainerID)
	w.WriteHeader(http.StatusNoContent)
}
