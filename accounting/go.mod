module github.com/thekndr/ates/accounting

go 1.20

replace github.com/thekndr/ates/common => ../common

require (
	github.com/confluentinc/confluent-kafka-go v1.9.2
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.10.9
	github.com/robfig/cron/v3 v3.0.1
	github.com/thekndr/ates/common v0.0.0-00010101000000-000000000000
)
