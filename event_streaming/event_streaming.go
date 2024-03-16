package event_streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
	"github.com/thekndr/ates/schema_registry"
	"log"
	"strconv"
)

type (
	EventStreaming struct {
		producer *kafka.Producer

		schemas       schema_registry.Schemas
		eventVersions EventVersions

		ctx    context.Context
		cancel func()
	}

	EventVersions = map[string]int

	EventContext = map[string]interface{}

	// For the service itself
	InternalEvent struct {
		Name    string
		Context map[string]interface{}
	}

	// For public clients
	PublicEvent struct {
		Name string `json:"name"`
		Meta struct {
			Id      string `json:"id"`
			Version string `json:"version"`
		} `json:"meta"`
		Context map[string]interface{} `json:"context"`
	}
)

func MustNewEventStreaming(schemas schema_registry.Schemas, eventVersions EventVersions) EventStreaming {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092"})
	if err != nil {
		log.Fatalf(`failed to create kafka producer: %s`, err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return EventStreaming{
		producer: producer,
		ctx:      ctx, cancel: cancel,
	}
}

func (es *EventStreaming) Cancel() {
	es.cancel()
}

func (es *EventStreaming) Start(topic string) chan InternalEvent {
	evCh := make(chan InternalEvent)
	go es.listenAndPublish(evCh, topic)
	return evCh
}

func (es *EventStreaming) Stop() {
	es.producer.Close()
}

func (es *EventStreaming) listenAndPublish(evCh <-chan InternalEvent, topic string) {
	for {
		select {
		case <-es.ctx.Done():
			fmt.Println("Shutting down goroutine...")
			return
		case internalEv := <-evCh:
			jsonEv, err := es.buildEvent(internalEv)
			if err != nil {
				log.Fatalf(`failed to build event: %w`, err)
				// TODO: error handling policy
				continue
			}

			err = es.producer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
				Value:          jsonEv,
			}, nil)
			if err != nil {
				log.Fatalf("Error producing to Kafka: %s\n", err)
				continue
			}
		}
	}
}

func (es *EventStreaming) buildEvent(ev InternalEvent) ([]byte, error) {
	evVersion := es.eventVersions[ev.Name]
	pubEv := rewrapEvent(ev, evVersion)

	jsonEv, err := json.Marshal(pubEv)
	if err != nil {
		return nil, fmt.Errorf(`failed to marshal ev=%=v to json: %w`, pubEv, err)
	}

	ok, err := es.schemas.Validate(jsonEv, ev.Name, evVersion)
	if err != nil {
		return nil, fmt.Errorf(`error durinv event=%s validation: %w`, pubEv.Name, err)
	}

	if !ok {
		return nil, fmt.Errorf(`the event=%s doesn't follow the schema version=%d`, ev.Name, evVersion)
	}

	return jsonEv, nil
}

func rewrapEvent(src InternalEvent, version int) PublicEvent {
	ev := PublicEvent{
		Name:    src.Name,
		Context: src.Context,
	}
	ev.Meta.Version = strconv.Itoa(version)
	ev.Meta.Id = uuid.NewString()

	return ev
}

func Publish(eventCh chan InternalEvent, name string, context map[string]interface{}) {
	go func() {
		eventCh <- InternalEvent{Name: name, Context: context}
	}()
}
