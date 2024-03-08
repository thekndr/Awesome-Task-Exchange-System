package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/thekndr/ates/task_management/users"
)

type ListTasks struct {
	Db       *sql.DB
	Managers *users.Managers
}

func (h *ListTasks) Handle(userId string, w http.ResponseWriter, r *http.Request) {
	type taskRow struct {
		CreatedAt   string `json:"created_at"`
		PublicId    string `json:"id"`
		Status      int    `json:"status"`
		AssigneeId  string `json:"assignee_id"`
		Description string `json:"description"`
	}

	var (
		query string
		err   error
		rows  *sql.Rows
	)
	if h.Managers.Has(userId) {
		query = `SELECT created_at, public_id, status, assignee_id, description FROM tasks;`
		rows, err = h.Db.Query(query)
		log.Printf(`- listing all tasks`)
	} else {
		query = `SELECT created_at, public_id, status, assignee_id, description FROM tasks WHERE assignee_id = $1;`
		rows, err = h.Db.Query(query, userId)
		log.Printf(`- listing tasks limited to user=%s`, userId)
	}

	if err != nil {
		http.Error(w, "(1) Failed to query assigned tasks", http.StatusInternalServerError)
		log.Printf(`failed to query assigned tasks, user-id=%s: %s`, userId, err)
		return
	}
	defer rows.Close()

	tasks := make([]taskRow, 0)

	for rows.Next() {
		var task taskRow
		if err := rows.Scan(&task.CreatedAt, &task.PublicId, &task.Status, &task.AssigneeId, &task.Description); err != nil {
			http.Error(w, "(2) Failed to query assigned tasks", http.StatusInternalServerError)
			log.Printf(`failed to scan queried rows, user-id=%s: %s`, userId, err)
			return
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "(3) Failed to query assigned tasks", http.StatusInternalServerError)
		log.Printf(`rows iteration error, user-id=%s: %s`, userId, err)
		return
	}

	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		http.Error(w, "Failed to encode tasks list as JSON", http.StatusInternalServerError)
		log.Printf(`failed to encode response user-id=%s, tasks=%+v: %s`, userId, tasks, err)
		return
	}
}
