package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/thekndr/ates/auth_client"
	"github.com/thekndr/ates/common"
	"github.com/thekndr/ates/event_streaming"
	"github.com/thekndr/ates/schema_registry"
	"github.com/thekndr/ates/task_management"
	"github.com/thekndr/ates/task_management/handlers"
	"github.com/thekndr/ates/task_management/users"
	"golang.org/x/exp/maps"
	"log"
	"net/http"
)

var (
	authAPIPort = 4000
	listenPort  = 4001
	db          = task_management.MustInitDB()

	workers     = users.NewWorkers()
	managers    = users.NewManagers()
	authSchemas = newAuthSchemas()

	authEventVersion  = 1
	selfEventVersions = event_streaming.EventVersions{
		"task-created":   1,
		"task-assigned":  1,
		"task-completed": 1,
	}
	selfEventTopic = "task-managements.tasks"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go mustConsumeFromKafka(ctx, "auth.users", func(msg []byte) error {
		return onAccountChange(msg, workers, managers)
	})

	kafkaStreaming := MustNewEventStreaming()
	defer kafkaStreaming.Cancel()
	eventCh := kafkaStreaming.Start(selfEventTopic)

	mux := establishEndpoints(eventCh)

	log.Printf("TaskManagement.Server started at port %d\n", listenPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", listenPort), mux))
}

func establishEndpoints(eventCh chan event_streaming.InternalEvent) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc(`GET /tasks`, auth_client.WithTokenVerification(
		authAPIPort, handlers.WithUserId(&handlers.ListTasks{
			Db: db, Managers: managers,
		}),
	))

	mux.HandleFunc(`PUT /tasks/new`, auth_client.WithTokenVerification(
		authAPIPort, handlers.WithManagerRoleOnly(&handlers.CreateTask{
			Db: db, Workers: workers, EventCh: eventCh,
		}),
	))

	mux.HandleFunc(`POST /tasks/shuffle`, auth_client.WithTokenVerification(
		authAPIPort, handlers.WithManagerRoleOnly(&handlers.ShuffleTasks{
			Db: db, Workers: workers, EventCh: eventCh,
		}),
	))

	mux.HandleFunc(`POST /tasks/complete/{taskId}`, auth_client.WithTokenVerification(
		authAPIPort, handlers.WithUserId(&handlers.CompleteTask{
			Db: db, EventCh: eventCh,
		}),
	))

	return mux
}

func onAccountChange(msg []byte, workers *users.Workers, managers *users.Managers) error {
	var ev event_streaming.PublicEvent

	if err := json.Unmarshal(msg, &ev); err != nil {
		return fmt.Errorf(`failed to decode event: %w`, err)
	}

	valid, err := authSchemas.Validate(msg, ev.Name, authEventVersion)
	if err != nil {
		log.Fatalf(`error during auth schema event=%s (%+v): %s`, ev.Name, ev.Meta, err)
	}

	if !valid {
		log.Fatalf(`invalid schema for event=%s (%+v)`, ev.Name, ev.Meta)
	}

	userIdStr := ev.Context["user-id"].(string)

	switch ev.Name {
	case "user-registered":
		log.Printf(`user added, id=%s`, userIdStr)

		role := common.Role(ev.Context["role"].(string))
		switch role {
		case common.ManagerRole:
			_ = managers.Add(userIdStr)
		case common.WorkerRole:
			_ = workers.Add(userIdStr,
				// for the sake of simplicity the type checking is omit
				ev.Context["email"].(string),
			)
		default:
			log.Fatalf(`unsupported role: %s`, role)
		}

	case "user-role-changed":
		// for the sake of simplicity type checking is omit
		newRole := common.Role(ev.Context["new-role"].(string))
		if !newRole.IsValid() {
			log.Printf(`invalid role received=%s`, newRole)
		} else {
			switch common.Role(newRole) {
			case common.WorkerRole:
				_ = workers.Add(userIdStr, "todo@email")
				_ = managers.Remove(userIdStr)
			case common.ManagerRole:
				_ = workers.Remove(userIdStr)
				_ = managers.Add(userIdStr)
			default:
				log.Fatalf(`unsupported role: %s`, newRole)
			}
			log.Printf(`role changed, id=%s, new-role=%s`, userIdStr, newRole)
		}
	}

	return nil
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

func MustNewEventStreaming() event_streaming.EventStreaming {
	schemas, err := schema_registry.NewSchemas(
		schema_registry.Scope("tasks"), maps.Keys(selfEventVersions)...,
	)
	if err != nil {
		log.Fatalf(`failed to create self schemas registry validator: %w`, err)
	}

	return event_streaming.MustNewEventStreaming(schemas, selfEventVersions)
}
