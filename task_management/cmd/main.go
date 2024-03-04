package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/thekndr/ates/auth_client"
	"github.com/thekndr/ates/common"
	"github.com/thekndr/ates/task_management"
	"github.com/thekndr/ates/task_management/handlers"
	"github.com/thekndr/ates/task_management/users"
	"log"
	"net/http"
)

var (
	authAPIPort = 4000
	listenPort  = 4001
	db          = task_management.MustInitDB()
	workers     = users.NewWorkers()
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go mustConsumeFromKafka(ctx, "accounts-stream", func(msg []byte) error {
		return onAccountChange(msg, workers)
	})
	mux := http.NewServeMux()

	mux.HandleFunc(`GET /tasks`, auth_client.WithTokenVerification(
		authAPIPort, handlers.WithUserId(&handlers.ListTasks{
			Db: db,
		}),
	))

	mux.HandleFunc(`PUT /tasks/new`, auth_client.WithTokenVerification(
		authAPIPort, handlers.WithManagerRoleOnly(&handlers.CreateTask{
			Db: db, Workers: workers,
		}),
	))

	mux.HandleFunc(`POST /tasks/shuffle`, auth_client.WithTokenVerification(
		authAPIPort, handlers.WithManagerRoleOnly(&handlers.ShuffleTasks{
			Db: db, Workers: workers,
		}),
	))

	mux.HandleFunc(`POST /tasks/complete/{taskId}`, auth_client.WithTokenVerification(
		authAPIPort, handlers.WithUserId(&handlers.CompleteTask{
			Db: db,
		}),
	))

	log.Printf("TaskManagement.Server started at port %d\n", listenPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", listenPort), mux))
}

func onAccountChange(msg []byte, workers *users.Workers) error {
	type event struct {
		Name    string                 `json:"event_name"`
		Context map[string]interface{} `json:"event_context"`
	}
	var ev event

	if err := json.Unmarshal(msg, &ev); err != nil {
		return fmt.Errorf(`failed to decode event: %w`, err)
	}

	userId, ok := ev.Context["user-id"]
	if !ok {
		return fmt.Errorf(`malformed event, user-id missing`)
	}
	userIdStr, ok := userId.(string)
	if !ok {
		return fmt.Errorf(`malformed event, user-id non-string type`)
	}

	switch ev.Name {
	case "user-registered":
		log.Printf(`user added, id=%s`, userIdStr)
		_ = workers.Add(
			userIdStr,
			// for the sake of simplicity type checking is omit
			ev.Context["email"].(string),
		)

	case "role-changed":
		// for the sake of simplicity type checking is omit
		newRole := common.Role(ev.Context["new-role"].(string))
		if !newRole.IsValid() {
			log.Printf(`invalid role received=%s`, newRole)
		} else {
			if common.Role(newRole) != common.WorkerRole {
				_ = workers.Remove(userIdStr)
				log.Printf(`non-worker user removed, id=%s, new-role=%s`, userIdStr, newRole)
			}
		}
	}

	return nil
}
