package main

import (
	"database/sql"
	"fmt"
	"github.com/thekndr/ates/auth_client"
	"github.com/thekndr/ates/common"
	"log"
	"net/http"
	"time"
)

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type Event struct {
	Name    string                 `json:"event_name"`
	Context map[string]interface{} `json:"event_context"`
}

var (
	listenPort    = MustGetEnvInt(`AUTH_API_PORT`)
	dashboardPort = MustGetEnvInt(`DASHBOARD_PORT`)

	db                       *sql.DB
	jwtToken                 = JwtToken{Secret: "s3cret jwt key", TTL: 24 * time.Hour}
	jwtKey                   = []byte("s3cret jwt key")
	loginRedirectUrlTemplate = "http://localhost:{{ dashboard_port }}/token={{ token }}"
)

func ensureValidRole(role string, w http.ResponseWriter) bool {
	if common.Role(role).IsValid() {
		return true
	}

	http.Error(w, fmt.Sprintf(`Invalid role: %s`, role), http.StatusBadRequest)
	return false
}

func main() {
	initDB()

	kafkaStreaming := MustNewEventStreaming()
	defer kafkaStreaming.Cancel()

	eventCh := kafkaStreaming.Start("accounts-stream")

	mux := http.NewServeMux()

	signup := Signup{EventCh: eventCh}
	mux.HandleFunc(`POST /signup`, signup.Handle)

	mux.HandleFunc(`POST /login`, loginHandler)
	mux.HandleFunc(`GET /verify`, verifyTokenHandler)

	changeRole := ChangeRole{EventCh: eventCh}
	// For the sake of simplicity, `manager` role is not verified here but required
	mux.HandleFunc(`POST /changeRole`, auth_client.WithTokenVerification(
		listenPort, changeRole.Handle,
	))
	mux.HandleFunc(`GET /listUsers`, auth_client.WithTokenVerification(
		listenPort, listUsersHandler,
	))

	fmt.Printf("Auth.Server started at port %d\n", listenPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", listenPort), mux))
}
