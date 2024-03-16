package handlers

import (
	"database/sql"

	"encoding/json"
	"log"
	"net/http"

	"github.com/thekndr/ates/common"
	"github.com/thekndr/ates/event_streaming"
	"github.com/thekndr/ates/task_management/users"
)

type CreateTask struct {
	Db      *sql.DB
	Workers *users.Workers
	EventCh chan event_streaming.InternalEvent
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
	var (
		taskId string
		err    error
	)
	if taskId, err = createTask(h.Db, taskDescription, assigneeId); err != nil {
		http.Error(w, "Failed to add a new task", http.StatusInternalServerError)
		log.Printf(`failed: create task, description=%s, user-id=%s: %s`, requestData.Description, assigneeId, err)
		return
	}

	log.Printf(`Task created id=%s, assigned to %s`, taskId, assigneeId)
	w.WriteHeader(http.StatusCreated)

	event_streaming.Publish(
		h.EventCh, "task-created",
		event_streaming.EventContext{
			"assignee-id": assigneeId,
			"description": taskDescription,
			"id":          taskId,
		},
	)
}

func createTask(db *sql.DB, description string, assigneeID string) (string, error) {
	query := `INSERT INTO tasks (description, assignee_id, status) VALUES ($1, $2, $3) RETURNING public_id`
	var publicId string
	err := db.QueryRow(query, description, assigneeID, common.TaskActiveStatus).Scan(&publicId)
	return publicId, err
}
