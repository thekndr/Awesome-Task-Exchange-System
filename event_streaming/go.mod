module github.com/thekndr/ates/event_streaming

replace github.com/thekndr/ates/schema_registry => ../schema_registry

go 1.20

require (
	github.com/confluentinc/confluent-kafka-go v1.9.2
	github.com/google/uuid v1.6.0
	github.com/thekndr/ates/schema_registry v0.0.0-00010101000000-000000000000
)

require (
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
)
