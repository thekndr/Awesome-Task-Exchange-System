package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

type ListTasks struct {
	Db *sql.DB
}

// TODO: list all tasks for non-worker roles
func (h *ListTasks) Handle(userId string, w http.ResponseWriter, r *http.Request) {
	log.Printf("tasks listed, user=%s", userId)

	type taskRow struct {
		CreatedAt   string `json:"created_at"`
		UUID        string `json:"id"`
		Status      int    `json:"status"`
		Description string `json:"description"`
	}

	query := `SELECT created_at, uuid, status, description FROM tasks WHERE assignee = $1;`
	rows, err := h.Db.Query(query, userId)
	if err != nil {
		http.Error(w, "(1) Failed to query assigned tasks", http.StatusInternalServerError)
		log.Printf(`failed to query assigned tasks, user-id=%s: %s`, userId, err)
		return
	}
	defer rows.Close()

	tasks := make([]taskRow, 0)

	for rows.Next() {
		var task taskRow
		if err := rows.Scan(&task.CreatedAt, &task.UUID, &task.Status, &task.Description); err != nil {
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
