package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
)

type EventStreaming struct {
	producer *kafka.Producer

	ctx    context.Context
	cancel func()
}

func MustNewEventStreaming() EventStreaming {
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

func (es *EventStreaming) Start(topic string) chan interface{} {
	evCh := make(chan interface{})
	go es.listenAndPublish(evCh, topic)
	return evCh
}

func (es *EventStreaming) Stop() {
	es.producer.Close()
}

func (es *EventStreaming) listenAndPublish(evCh <-chan interface{}, topic string) {
	for {
		select {
		case <-es.ctx.Done():
			fmt.Println("Shutting down goroutine...")
			return
		case data := <-evCh:
			jsonData, err := json.Marshal(data)
			if err != nil {
				fmt.Printf("Error marshalling JSON: %s\n", err)
				continue
			}
			err = es.producer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
				Value:          jsonData,
			}, nil)
			if err != nil {
				log.Fatalf("Error producing to Kafka: %s\n", err)
				continue
			}
		}
	}
}
