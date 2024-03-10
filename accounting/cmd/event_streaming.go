package main

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
	"log"
)

type OnKafkaMessageFunc func(topic string, payload []byte) error

func mustConsumeFromKafka(ctx context.Context, topics []string, onMessage OnKafkaMessageFunc) {
	uniqueGroupId := fmt.Sprintf(`ates-accounting-%s`, uuid.NewString())

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  "localhost:9092",
		"group.id":           uniqueGroupId,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": true,
	})

	if err != nil {
		log.Fatalf("Failed to create consumer: %s", err)
	}

	defer c.Close()

	err = c.SubscribeTopics(topics, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to topics: %s", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down consumer goroutine")
			return
		default:
			msg, err := c.ReadMessage(-1)
			if err == nil {
				go func() {
					if err := onMessage(*msg.TopicPartition.Topic, msg.Value); err != nil {
						log.Printf(`on-message handler failed: %s`, err)
					}
				}()
				// Note: We're not committing the offset.
			} else {
				// The client will automatically try to recover from all errors.
				log.Printf("Consumer error: %v (%v)\n", err, msg)
			}
		}
	}
}
