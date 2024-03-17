package main

import (
	"fmt"
	"github.com/thekndr/ates/analytics/event_handlers"
	"github.com/thekndr/ates/event_streaming"
	"github.com/thekndr/ates/schema_registry"
	"log"
	"time"
)

var (
	accountingSchemas      = newAccountingSchemas()
	accountingEventVersion = 1
)

type eventHandlers struct {
	task struct {
		assigned  event_handlers.TaskAssigned
		completed event_handlers.TaskCompleted
	}
}

func (eh *eventHandlers) OnEvent(_ string, ev event_streaming.PublicEvent, rawEv []byte) error {
	if err := eh.onTaskAccountingEvent(ev, rawEv); err != nil {
		return fmt.Errorf(`error during handling task account event: %w`, err)
	}

	return nil
}

func (eh *eventHandlers) onTaskAccountingEvent(ev event_streaming.PublicEvent, rawEv []byte) error {
	if err := eh.validate(ev, rawEv); err != nil {
		return err
	}

	taskTime := ev.Context["time"].(time.Time)
	taskId := ev.Context["id"].(string)
	assigneeId := ev.Context["assignee-id"].(string)

	var err error

	switch ev.Name {
	case "task-assigned":
		taskCost := ev.Context["cost"].(uint)
		err = eh.task.assigned.Handle(event_handlers.TaskAssignedEvent{
			Time: taskTime, TaskId: taskId,
			WorkerId: assigneeId, TaskCost: taskCost,
		})

	case "task-completed":
		taskReward := ev.Context["reward"].(uint)
		err = eh.task.completed.Handle(event_handlers.TaskCompletedEvent{
			Time: taskTime, TaskId: taskId,
			WorkerId: assigneeId, TaskReward: taskReward,
		})
	}

	return err
}

func (eh *eventHandlers) validate(ev event_streaming.PublicEvent, rawEv []byte) error {
	valid, err := accountingSchemas.Validate(rawEv, ev.Name, accountingEventVersion)
	if err != nil {
		return fmt.Errorf(`error during schema event=%s (%+v): %s`, ev.Name, ev.Meta, err)
	}

	if !valid {
		return fmt.Errorf(`validation for event=%s versions=%d failed`, ev.Name, accountingEventVersion)
	}

	return nil
}

func newAccountingSchemas() schema_registry.Schemas {
	authSchemas, err := schema_registry.NewSchemas(
		schema_registry.Scope("accounting"),
		"task-assigned", "task-completed",
	)
	if err != nil {
		log.Fatalf(`failed to create task schemas registry validator: %w`, err)
	}
	return authSchemas
}
