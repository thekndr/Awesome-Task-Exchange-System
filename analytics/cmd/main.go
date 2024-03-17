package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/thekndr/ates/analytics/event_handlers"
	"github.com/thekndr/ates/analytics/states"
	"github.com/thekndr/ates/event_streaming"
	"log"
	"net/http"
	"time"
)

var (
	selfApiListenPort = 4002

	companyBalanceState = states.CompanyBalance{}
	workerBalanceState  = states.WorkerBalance{}
	tasksState          = states.Tasks{}
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	topics := []string{"accounting.tasks"}
	log.Printf(`Listening to events (%s)...`, topics)

	eh := event_streaming.MustNewEventHandling(event_streaming.EventHandlingConfig{
		EnableAutoCommit: true,
	})

	evHandlers := newEventHandles()
	go func() {
		if err := eh.StartSync(ctx, topics, evHandlers.OnEvent); err != nil {
			log.Fatal(err)
		}
	}()

	mustRunHttpEndpoint()
}

func mustRunHttpEndpoint() {
	mux := http.NewServeMux()
	mux.HandleFunc(`GET /stat/profit/today`, func(w http.ResponseWriter, r *http.Request) {
		var response struct {
			CompanyProfit              int `json:"company-profit"`
			WorkersWithNegativeBalance int `json:"workers-with-negative-balance"`
		}
		response.CompanyProfit = companyBalanceState.ProfitFor(time.Now())
		response.WorkersWithNegativeBalance = workerBalanceState.WorkersWithNegativeBalance(time.Now())

		w.Header().Set(`Content-Type`, `application/json`)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response as JSON", http.StatusInternalServerError)
		}
	})

	// TODO: other endpoints
	mux.HandleFunc(`GET /stat/user-expensive-task/today`, func(_ http.ResponseWriter, _ *http.Request) {
		// ...
		_, _, _ = tasksState.MostExpensiveUserTask(time.Now())
		///
	})

	log.Printf(`Server started at port %d`, selfApiListenPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(`:%d`, selfApiListenPort), mux))
}

func newEventHandles() eventHandlers {
	evHandlers := eventHandlers{}
	evHandlers.task.assigned = event_handlers.TaskAssigned{
		CompanyBalance: &companyBalanceState,
		WorkerBalance:  &workerBalanceState,
		Tasks:          &tasksState,
	}
	evHandlers.task.completed = event_handlers.TaskCompleted{
		CompanyBalance: &companyBalanceState,
		WorkerBalance:  &workerBalanceState,
	}

	return evHandlers
}
