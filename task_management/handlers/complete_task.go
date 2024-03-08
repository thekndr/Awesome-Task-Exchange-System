package handlers

import (
	"database/sql"
	"github.com/thekndr/ates/common"
	"log"
	"net/http"
)

type CompleteTask struct {
	Db *sql.DB
}

func (h *CompleteTask) Handle(userId string, w http.ResponseWriter, r *http.Request) {
	taskId := r.PathValue(`taskId`)
	log.Printf("complete task=%s, user=%s", taskId, userId)

	query := `UPDATE tasks SET status = $1 WHERE public_id = $2 AND assignee_id = $3;`

	result, err := h.Db.Exec(query, common.TaskDoneStatus, taskId, userId)
	if err != nil {
		http.Error(w, "(1) Failed to complete the specified task", http.StatusInternalServerError)
		log.Printf(`failed to exec query (complete) task-id=%s user-id=%s: %s`, taskId, userId, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "(2) Failed to complete the specified task", http.StatusInternalServerError)
		log.Printf(`(complete) rows-affected task-id=%s user-id=%s: %d, %s`, taskId, userId, rowsAffected, err)
	}

	w.WriteHeader(http.StatusAccepted)
}
