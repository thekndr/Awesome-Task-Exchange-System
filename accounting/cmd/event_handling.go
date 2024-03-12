package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/thekndr/ates/accounting/db"
	"github.com/thekndr/ates/accounting/event_handlers"
	"github.com/thekndr/ates/common"
	"log"
)

type eventHandlers struct {
	account struct {
		created     event_handlers.AccountCreated
		roleChanged event_handlers.AccountRoleChanged
	}

	task struct {
		created   event_handlers.TaskCreated
		assigned  event_handlers.TaskAssigned
		completed event_handlers.TaskCompleted
	}

	billingCycleCompleted event_handlers.BillingCycleCompleted
}

func (eh *eventHandlers) setup(dbInstance *sql.DB) {
	workers := db.Workers{Db: dbInstance}
	tasks := db.Tasks{Db: dbInstance}
	transactions := db.Transactions{Db: dbInstance}
	billingCycles := db.BillingCycles{Db: dbInstance}

	eh.account.created = event_handlers.AccountCreated{Workers: workers}
	eh.account.roleChanged = event_handlers.AccountRoleChanged{}

	eh.task.created = event_handlers.TaskCreated{Tasks: tasks}
	eh.task.assigned = event_handlers.TaskAssigned{
		Transactions:  transactions,
		Tasks:         tasks,
		BillingCycles: billingCycles,
		Workers:       workers,
	}
	eh.task.completed = event_handlers.TaskCompleted{
		Transactions:  transactions,
		Tasks:         tasks,
		BillingCycles: billingCycles,
		Workers:       workers,
	}

	eh.billingCycleCompleted = event_handlers.BillingCycleCompleted{
		BillingCycles: billingCycles, Workers: workers,
	}
}

type event struct {
	Name    string                 `json:"event_name"`
	Context map[string]interface{} `json:"event_context"`
}

func (eh *eventHandlers) OnBillingCycleCompleted() error {
	workerIdsWithPositiveBalances, err := eh.billingCycleCompleted.Handle(event_handlers.BillingCycleCompletedEvent{})
	_ = workerIdsWithPositiveBalances

	// TODO:
	// - create payments
	// - trigger external PaymentSystem
	// - update payments (status=complete) AND reset balances
	return err
}

func (eh *eventHandlers) OnEvent(topic string, msg []byte) error {
	var ev event

	if err := json.Unmarshal(msg, &ev); err != nil {
		return fmt.Errorf(`failed to decode event: %w`, err)
	}

	switch topic {
	case "auth.accounts":
		return eh.onAuthEvent(ev)
	case "task-management.tasks":
		return eh.onTaskEvent(ev)
	default:
		log.Fatalf(`unknown/unsupported topic: %s`, topic)
	}
	return nil
}

func (eh *eventHandlers) onAuthEvent(ev event) error {
	userId, ok := ev.Context["user-id"]
	if !ok {
		return fmt.Errorf(`malformed event, user-id missing`)
	}
	userIdStr, ok := userId.(string)
	if !ok {
		return fmt.Errorf(`malformed event, user-id non-string type`)
	}

	switch ev.Name {
	case "user.registered":
		log.Printf(`user added, id=%s`, userIdStr)
		role := common.Role(ev.Context["role"].(string))

		switch role {
		case common.ManagerRole:
			log.Printf(`Manager role is ignored for now`)
		case common.WorkerRole:
			// TODO: error checking
			return eh.account.created.Handle(event_handlers.WorkerAccountCreatedEvent{
				Id:    userIdStr,
				Email: ev.Context["email"].(string),
			})

		default:
			log.Fatalf(`unsupported role: %s`, role)
		}

	case "user.role-changed":
		return eh.account.roleChanged.Handle(event_handlers.AccountRoleChangedEvent{
			Id:      userIdStr,
			Email:   ev.Context["email"].(string),
			NewRole: ev.Context["new-role"].(string),
		})

	default:
		log.Fatalf(`unsupported/unknown event: %s`, ev.Name)
	}

	return nil
}

func (eh *eventHandlers) onTaskEvent(ev event) error {
	taskId, ok := ev.Context["task-id"]
	if !ok {
		return fmt.Errorf(`malformed event, task-id missing`)
	}
	taskIdStr, ok := taskId.(string)
	if !ok {
		return fmt.Errorf(`malformed event, task-id non-string type`)
	}

	switch ev.Name {
	case "task.created":
		return eh.task.created.Handle(event_handlers.TaskCreatedEvent{
			Id:          taskIdStr,
			Description: ev.Context["description"].(string),
		})

	case "task.assigned":
		return eh.task.assigned.Handle(event_handlers.TaskAssignedEvent{
			Id:         taskIdStr,
			AssigneeId: ev.Context["assignee-id"].(string),
		})

	case "task.completed":
		return eh.task.completed.Handle(event_handlers.TaskCompletedEvent{
			Id:         taskIdStr,
			AssigneeId: ev.Context["assignee-id"].(string),
		})

	default:
		log.Fatalf(`unsupported/unknown event: %s`, ev.Name)
	}

	return nil
}
