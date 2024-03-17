package main

import (
	"context"
	cronSys "github.com/robfig/cron/v3"
	"github.com/thekndr/ates/accounting/db"
	"github.com/thekndr/ates/event_streaming"
	"github.com/thekndr/ates/schema_registry"
	"golang.org/x/exp/maps"
	"log"
	"os"
)

var (
	selfEventVersions = event_streaming.EventVersions{
		"task-created":   1,
		"task-completed": 1,
	}
	selfEventTopic = "accounting.tasks"
)

func main() {
	log.Printf(`Initializing db...`)
	dbInstance := db.MustInit(
		// shortcut for a pure environment
		os.Getenv(`ATES_ACCOUNTING_RESET_ALL_TABLES`) == "reset_all_accounting_tables",
	)

	kafkaStreaming := mustNewEventStreaming()
	defer kafkaStreaming.Cancel()

	eventCh := kafkaStreaming.Start(selfEventTopic)

	log.Printf(`configuring event handlers...`)
	var evHandlers eventHandlers
	evHandlers.setup(dbInstance, eventCh)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Printf(`Starting cron...`)
	cron := cronSys.New()
	cron.AddFunc("59 59 23 * * *", func() {
		log.Println(`Time to complete billing cycle`)
		evHandlers.OnBillingCycleCompleted()
	})
	cron.Start()

	topics := []string{"auth.accounts", "task-managements.tasks"}
	log.Printf(`Listening to events (%s)...`, topics)

	eh := event_streaming.MustNewEventHandling(event_streaming.EventHandlingConfig{
		EnableAutoCommit: true,
	})
	if err := eh.StartSync(ctx, topics, evHandlers.OnEvent); err != nil {
		log.Fatal(err)
	}
}

func mustNewEventStreaming() event_streaming.EventStreaming {
	schemas, err := schema_registry.NewSchemas(
		schema_registry.Scope("accounting"),
		maps.Keys(selfEventVersions)...,
	)
	if err != nil {
		log.Fatalf(`failed to create schemas registry validator: %w`, err)
	}

	return event_streaming.MustNewEventStreaming(schemas, selfEventVersions)
}
