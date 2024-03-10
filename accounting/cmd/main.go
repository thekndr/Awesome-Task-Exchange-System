package main

import (
	"context"
	cronSys "github.com/robfig/cron/v3"
	"github.com/thekndr/ates/accounting/db"
	"log"
	"os"
)

func main() {
	log.Printf(`Initializing db...`)
	dbInstance := db.MustInit(
		// shortcut for a pure environment
		os.Getenv(`ATES_ACCOUNTING_DROP_ALL_TABLES`) == "drop_all_accounting_tables",
	)
	- = dbInstance

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Printf(`Starting cron...`)
	cron := cronSys.New()
	cron.AddFunc("59 59 23 * * *", func() {
		log.Println(`Time to complete billing cycle`)
	})
	cron.Start()

	topics := []string{"auth.accounts", "task-managements.tasks"}
	log.Printf(`Listening to events (%s)...`, topics)
	mustConsumeFromKafka(ctx, topics, nil)
}
