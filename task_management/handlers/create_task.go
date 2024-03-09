package handlers

import (
	"database/sql"

	"encoding/json"
	"log"
	"net/http"

	"github.com/thekndr/ates/common"
	"github.com/thekndr/ates/task_management/users"
)

type CreateTask struct {
	Db      *sql.DB
	Workers *users.Workers
}

func (h *CreateTask) Handle(w http.ResponseWriter, r *http.Request) {
	type createTaskRequest struct {
		Description string `json:"description"`
	}

	var requestData createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf(`failed to decode new-task request (%v): %s`, r.Body, err)
		return
	}

	randomWorkerIds := h.Workers.RandomIds()
	if randomWorkerIds.Len() == 0 {
		http.Error(w, "There are no active users", http.StatusInternalServerError)
		log.Printf(`no active users`)
		return
	}

	taskDescription, assigneeId := requestData.Description, randomWorkerIds.Get()
	if err := createTask(h.Db, taskDescription, assigneeId); err != nil {
		http.Error(w, "Failed to add a new task", http.StatusInternalServerError)
		log.Printf(`failed: create task, description=%s, user-id=%s: %s`, requestData.Description, assigneeId, err)
		return
	}

	log.Printf(`Task created, assigned to %s`, assigneeId)
	w.WriteHeader(http.StatusCreated)
}

func createTask(db *sql.DB, description string, assigneeID string) error {
	query := `INSERT INTO tasks (description, assignee_id, status) VALUES ($1, $2, $3)`
	_, err := db.Exec(query, description, assigneeID, common.TaskActiveStatus)
	return err
}
