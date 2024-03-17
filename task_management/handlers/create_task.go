package handlers

import (
	"database/sql"

	"encoding/json"
	"log"
	"net/http"
	"regexp"

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
		JiraId      string `json:"jira_id"`
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

	taskDescription, assigneeId := parseTaskDescription(requestData.Description), randomWorkerIds.Get()
	if requestData.JiraId != "" {
		log.Printf(`overwriting jira_id from description="%s" to "%s" from request`, requestData.Description, requestData.JiraId)
		taskDescription.JiraId = requestData.JiraId
	}

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

	// NOTE/TODO: since the new task cannot be left unassigned, the assignment happens on creation
	// For the sake of simplicity, a single event `task-created` with assignment info is emited;
	// this could be splitted to two separate sequential events `created` + `assigned`.
	event_streaming.Publish(
		h.EventCh, "task-created",
		event_streaming.EventContext{
			"assignee-id": assigneeId,
			"description": taskDescription.Body,
			"jira-id":     taskDescription.JiraId,
			"id":          taskId,
		},
	)
}

func createTask(db *sql.DB, description parsedTaskDescription, assigneeID string) (string, error) {
	query := `INSERT INTO tasks (jira_id, description, assignee_id, status) VALUES ($1, $2, $3, $4) RETURNING public_id`
	var publicId string
	err := db.QueryRow(query, description.JiraId, description.Body, assigneeID, common.TaskActiveStatus).Scan(&publicId)
	return publicId, err
}

var (
	taskJiraTitleRe = `^\[(.*?)\] - (.*)$`
)

type parsedTaskDescription struct {
	JiraId string
	Body   string
}

func parseTaskDescription(input string) parsedTaskDescription {
	re := regexp.MustCompile(taskJiraTitleRe)
	matches := re.FindStringSubmatch(input)
	if len(matches) > 0 {
		return parsedTaskDescription{JiraId: matches[1], Body: matches[2]}
	}

	return parsedTaskDescription{Body: input}
}
