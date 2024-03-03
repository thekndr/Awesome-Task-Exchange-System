package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/thekndr/ates/common"
	"github.com/thekndr/ates/task_management/users"
)

type ShuffleTasks struct {
	Db      *sql.DB
	Workers *users.Workers
}

func (h *ShuffleTasks) Handle(w http.ResponseWriter, r *http.Request) {
	allWorkerIds := h.Workers.AllIds()

	changes, err := shuffleTaskAssignees(h.Db, allWorkerIds)
	if err != nil {
		http.Error(w, "Failed to shuffle tasks", http.StatusInternalServerError)
		log.Printf(`failed to shuffle tasks: %s`, err)
		return
	}

	if err := json.NewEncoder(w).Encode(changes); err != nil {
		http.Error(w, "Failed to encode changes as JSON", http.StatusInternalServerError)
		log.Printf(`failed to encode changes changes=%+v: %s`, changes, err)
		return
	}
}

type shuffleChange struct {
	TaskId        string `json:"task_id"`
	OldAssigneeId string `json:"old_assignee_id"`
	NewAssigneeId string `json:"new_assignee_id"`
}

func shuffleTaskAssignees(db *sql.DB, allWorkerIds []string) ([]shuffleChange, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query("SELECT assigneeId, id FROM tasks WHERE status = $1", common.TaskActiveStatus)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	defer rows.Close()

	var changes []shuffleChange
	for rows.Next() {
		var id int
		var oldAssigneeID string
		if err := rows.Scan(&oldAssigneeID, &id); err != nil {
			tx.Rollback()
			return nil, err
		}

		newAssigneeID := users.Random(allWorkerIds)
		_, err := tx.Exec("UPDATE tasks SET assignee_id = $1 WHERE id = $2", newAssigneeID, id)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		changes = append(changes, shuffleChange{
			TaskId: strconv.Itoa(id), OldAssigneeId: oldAssigneeID, NewAssigneeId: newAssigneeID,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return changes, nil
}
