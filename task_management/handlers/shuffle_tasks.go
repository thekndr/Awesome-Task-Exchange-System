package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/thekndr/ates/common"
	"github.com/thekndr/ates/task_management/users"
)

type ShuffleTasks struct {
	Db      *sql.DB
	Workers *users.Workers
}

func (h *ShuffleTasks) Handle(w http.ResponseWriter, r *http.Request) {
	changes, err := shuffleTaskAssignees(h.Db, h.Workers.RandomIds())
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

func shuffleTaskAssignees(db *sql.DB, randomWorkerIds users.RandomWorkerIds) ([]shuffleChange, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf(`failed to start tx: %w`, err)
	}

	rows, err := tx.Query("SELECT assignee_id, public_id FROM tasks WHERE status = $1", common.TaskActiveStatus)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type task struct {
		publicId   string
		assigneeId string
	}
	tasks := make([]task, 0)
	for rows.Next() {
		var t task
		if err := rows.Scan(&t.assigneeId, &t.publicId); err != nil {
			tx.Rollback()
			return nil, err
		}
		tasks = append(tasks, t)
	}

	var changes []shuffleChange
	for _, t := range tasks {
		var (
			oldAssigneeId string = t.assigneeId
			newAssigneeId string = oldAssigneeId
		)
		if randomWorkerIds.Len() > 1 {
			for ; newAssigneeId == oldAssigneeId; newAssigneeId = randomWorkerIds.Get() {
				// log.Printf("looking for a new random assigneed, all: %+v", randomWorkerIds)
			}
		}

		if newAssigneeId == "" {
			changes = append(changes,
				shuffleChange{TaskId: t.publicId, OldAssigneeId: oldAssigneeId, NewAssigneeId: oldAssigneeId},
			)
			log.Printf(`shuffle: not enough alternatives, nothing has changed actually`)
			continue
		}

		_, err := tx.Exec("UPDATE tasks SET assignee_id = $1 WHERE public_id = $2", newAssigneeId, t.publicId)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		changes = append(changes, shuffleChange{
			TaskId: t.publicId, OldAssigneeId: oldAssigneeId, NewAssigneeId: newAssigneeId,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return changes, nil
}
