package main

import (
	"database/sql"
	"fmt"
	"github.com/thekndr/ates/auth_client"
	"github.com/thekndr/ates/common"
	"github.com/thekndr/ates/event_streaming"
	"github.com/thekndr/ates/schema_registry"
	"golang.org/x/exp/maps"
	"log"
	"net/http"
	"time"
)

type (
	User struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
)

var (
	listenPort    = MustGetEnvInt(`AUTH_API_PORT`)
	dashboardPort = MustGetEnvInt(`DASHBOARD_PORT`)

	db                       *sql.DB
	jwtToken                 = JwtToken{Secret: "s3cret jwt key", TTL: 24 * time.Hour}
	jwtKey                   = []byte("s3cret jwt key")
	loginRedirectUrlTemplate = "http://localhost:{{ dashboard_port }}/token={{ token }}"

	selfEventVersions = event_streaming.EventVersions{
		"user-registered":   1,
		"user-role-changed": 1,
	}
	eventTopic = "auth.users"
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

	eventCh := kafkaStreaming.Start(eventTopic)

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

func MustNewEventStreaming() event_streaming.EventStreaming {
	schemas, err := schema_registry.NewSchemas(
		schema_registry.Scope("auth"),
		maps.Keys(selfEventVersions)...,
	)
	if err != nil {
		log.Fatalf(`failed to create schemas registry validator: %w`, err)
	}

	return event_streaming.MustNewEventStreaming(schemas, selfEventVersions)
}
