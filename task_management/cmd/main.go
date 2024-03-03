package main

import (
	"context"
	"fmt"
	"github.com/thekndr/ates/auth_client"
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
