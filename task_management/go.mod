module github.com/thekndr/ates/task_management

go 1.22

replace github.com/thekndr/ates/auth_client => ../auth_client

replace github.com/thekndr/ates/common => ../common

replace github.com/thekndr/ates/event_streaming => ../event_streaming

replace github.com/thekndr/ates/schema_registry => ../schema_registry

require (
	github.com/confluentinc/confluent-kafka-go v1.9.2
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.10.9
	github.com/thekndr/ates/auth_client v0.0.0-00010101000000-000000000000
	github.com/thekndr/ates/common v0.0.0-00010101000000-000000000000
	github.com/thekndr/ates/event_streaming v0.0.0-00010101000000-000000000000
	github.com/thekndr/ates/schema_registry v0.0.0-00010101000000-000000000000
	golang.org/x/exp v0.0.0-20240314144324-c7f7c6466f7f
)

require (
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
)
