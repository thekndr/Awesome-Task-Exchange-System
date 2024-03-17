package main

import (
	"database/sql"
	"fmt"
	"github.com/thekndr/ates/accounting/db"
	"github.com/thekndr/ates/accounting/event_handlers"
	"github.com/thekndr/ates/common"
	"github.com/thekndr/ates/event_streaming"
	"github.com/thekndr/ates/schema_registry"
	"log"
)

var (
	authSchemas      = newAuthSchemas()
	authEventVersion = 1

	taskSchemas       = newTaskSchemas()
	taskEventVersions = []int{1, 2}
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

func (eh *eventHandlers) setup(dbInstance *sql.DB, eventCh chan event_streaming.InternalEvent) {
	workers := db.Workers{Db: dbInstance}
	tasks := db.Tasks{Db: dbInstance}
	transactions := db.Transactions{Db: dbInstance}
	billingCycles := db.BillingCycles{Db: dbInstance}

	eh.account.created = event_handlers.AccountCreated{Workers: workers}
	eh.account.roleChanged = event_handlers.AccountRoleChanged{}

	eh.task.created = event_handlers.TaskCreated{Tasks: tasks}
	eh.task.assigned = event_handlers.TaskAssigned{
		EventCh:       eventCh,
		Transactions:  transactions,
		Tasks:         tasks,
		BillingCycles: billingCycles,
		Workers:       workers,
	}
	eh.task.completed = event_handlers.TaskCompleted{
		EventCh:       eventCh,
		Transactions:  transactions,
		Tasks:         tasks,
		BillingCycles: billingCycles,
		Workers:       workers,
	}

	eh.billingCycleCompleted = event_handlers.BillingCycleCompleted{
		BillingCycles: billingCycles, Workers: workers,
	}
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

func (eh *eventHandlers) OnEvent(topic string, ev event_streaming.PublicEvent, rawEv []byte) error {
	switch topic {
	case "auth.accounts":
		return eh.onAuthEvent(ev, rawEv)
	case "task-management.tasks":
		return eh.onTaskEvent(ev, rawEv)
	default:
		log.Fatalf(`unknown/unsupported topic: %s`, topic)
	}
	return nil
}

func (eh *eventHandlers) onAuthEvent(ev event_streaming.PublicEvent, rawEv []byte) error {
	if _, err := eh.validate(authSchemas, ev, rawEv, authEventVersion); err != nil {
		return err
	}

	userIdStr := ev.Context["user-id"].(string)
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

func (eh *eventHandlers) onTaskEvent(ev event_streaming.PublicEvent, rawEv []byte) error {
	var (
		schemaVersion int
		err           error
	)
	if schemaVersion, err = eh.validate(taskSchemas, ev, rawEv, taskEventVersions...); err != nil {
		return err
	}

	taskIdStr := ev.Context["id"].(string)
	switch ev.Name {
	case "task-created":
		// Though `JiraId` is not used anywhere for accounting logic,
		// this is an example of possible version handling
		var jiraId string
		if schemaVersion == 2 {
			jiraId = ev.Context["jira-id"].(string)
		}

		return eh.task.created.Handle(event_handlers.TaskCreatedEvent{
			Id:          taskIdStr,
			Description: ev.Context["description"].(string),
			JiraId:      jiraId,
		})

	case "task-assigned":
		return eh.task.assigned.Handle(event_handlers.TaskAssignedEvent{
			Id:         taskIdStr,
			AssigneeId: ev.Context["assignee-id"].(string),
		})

	case "task-completed":
		return eh.task.completed.Handle(event_handlers.TaskCompletedEvent{
			Id:         taskIdStr,
			AssigneeId: ev.Context["assignee-id"].(string),
		})

	default:
		log.Fatalf(`unsupported/unknown event: %s`, ev.Name)
	}

	return nil
}

func (eg *eventHandlers) validate(schemas schema_registry.Schemas, ev event_streaming.PublicEvent, rawEv []byte, versions ...int) (int, error) {
	for _, ver := range versions {
		valid, err := schemas.Validate(rawEv, ev.Name, ver)
		if err != nil {
			return -1, fmt.Errorf(`error during schema event=%s (%+v): %s`, ev.Name, ev.Meta, err)
		}

		if valid {
			return ver, nil
		}
	}

	return -1, fmt.Errorf(`validation for event=%s versions=%+v failed`, ev.Name, versions)
}

func newAuthSchemas() schema_registry.Schemas {
	authSchemas, err := schema_registry.NewSchemas(
		schema_registry.Scope("auth"),
		"user-registered", "user-role-changed",
	)
	if err != nil {
		log.Fatalf(`failed to create auth schemas registry validator: %w`, err)
	}
	return authSchemas
}

func newTaskSchemas() schema_registry.Schemas {
	authSchemas, err := schema_registry.NewSchemas(
		schema_registry.Scope("task"),
		"task-created", "task-assigned", "task-completed",
	)
	if err != nil {
		log.Fatalf(`failed to create task schemas registry validator: %w`, err)
	}
	return authSchemas
}
