package event_handlers

import (
	"database/sql"
	"fmt"
	"github.com/thekndr/ates/accounting/db"
)

type (
	TaskAssignedEvent struct {
		Id         string `json:"id"`
		AssigneeId string `json:"assignee_id"`
	}

	TaskAssigned struct {
		Transactions  db.Transactions
		Tasks         db.Tasks
		BillingCycles db.BillingCycles
		Workers       db.Workers
	}
)

func (h *TaskAssigned) Handle(ev TaskAssignedEvent) error {
	accTask, err := h.Tasks.GetByPublic(ev.Id)
	if err != nil {
		return fmt.Errorf(`task-assigned: failed to find a task=%s: %w`, ev.Id, err)
	}

	tx, err := h.BillingCycles.Tx()
	if err != nil {
		return err
	}

	current, err := h.BillingCycles.CurrentOrNew(tx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf(`task-assigned: failed to get/create current billing-cycle: %w`, err)
	}

	taskId, workerId, err := convertIds(h.Tasks, h.Workers, ev.Id, ev.AssigneeId)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf(`task-assigned: failed to map ids: %w`, err)
	}

	accTx, err := h.Transactions.New(current.Id, taskId, workerId)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf(`task-assigned: failed to find/create acc-tx for task=%s`, ev.Id)
	}

	if err := accTx.Withdraw(tx, accTask.AssignmentPrice); err != nil {
		tx.Rollback()
		return fmt.Errorf(`task-assigned: failed to withdraw user=%s task=%s: %w`, ev.AssigneeId, ev.Id, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf(`task-assigned: failed to commit db-tx`)
	}

	return nil
}

func convertIds(tasks db.Tasks, workers db.Workers, publicTaskId, publicUserId string) (taskId, userId uint64, err error) {
	var task db.Task
	task, err = tasks.GetByPublic(publicTaskId)
	if err != nil {
		return
	}

	var worker db.Worker
	worker, err = workers.GetByPublic(publicUserId)
	if err != nil {
		return
	}

	return task.Id, worker.Id, nil
}

type (
	TaskCompletedEvent struct {
		Id         string `json:"id"`
		AssigneeId string `json:"id"`
	}

	TaskCompleted struct {
		Db            *sql.DB
		Transactions  db.Transactions
		Tasks         db.Tasks
		BillingCycles db.BillingCycles
		Workers       db.Workers
	}
)

func (h *TaskCompleted) Handle(ev TaskCompletedEvent) error {
	accTask, err := h.Tasks.GetByPublic(ev.Id)
	if err != nil {
		return fmt.Errorf(`task-assigned: failed to find a task=%s: %w`, ev.Id, err)
	}

	tx, err := h.BillingCycles.Tx()
	if err != nil {
		return err
	}

	current, err := h.BillingCycles.Current(tx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf(`task-assigned: failed to get/create current billing-cycle: %w`, err)
	}

	taskId, workerId, err := convertIds(h.Tasks, h.Workers, ev.Id, ev.AssigneeId)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf(`task-assigned: failed to map ids: %w`, err)
	}

	accTx, err := h.Transactions.Get(current.Id, taskId, workerId)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf(`task-assigned: failed to find acc-tx for task=%s`, ev.Id)
	}

	if err := accTx.Enroll(tx, accTask.RewardPrice); err != nil {
		tx.Rollback()
		return fmt.Errorf(`task-assigned: failed to withdraw user=%s task=%s: %w`, ev.AssigneeId, ev.Id, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf(`task-assigned: failed to commit db-tx`)
	}

	return nil
}

type (
	TaskCreatedEvent struct {
		Id          string `json:"id"`
		JiraId      string `json:"jira_id"`
		Description string `json:"description"`
	}

	TaskCreated struct {
		Tasks db.Tasks
	}
)

func (h *TaskCreated) Handle(ev TaskCreatedEvent) error {
	_, err := h.Tasks.Add(ev.Id, ev.Description)
	return err
}
