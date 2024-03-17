package event_streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
	"log"
)

type (
	OnKafkaMessageFunc func(topic string, event PublicEvent, rawEvent []byte) error

	EventHandlingConfig struct {
		// unique when is omitted
		GroupId string

		BootstrapServer  string
		EnableAutoCommit bool
	}

	EventHandling struct {
		c *kafka.Consumer
	}
)

func MustNewEventHandling(config EventHandlingConfig) EventHandling {
	if config.GroupId == "" {
		config.GroupId = fmt.Sprintf(`ates-%s`, uuid.NewString())
	}
	if config.BootstrapServer == "" {
		config.BootstrapServer = "localhost:9092"
	}

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  config.BootstrapServer,
		"group.id":           config.GroupId,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": config.EnableAutoCommit,
	})

	if err != nil {
		log.Fatalf("Failed to create consumer: %s", err)
	}

	return EventHandling{c: c}
}

func (eh *EventHandling) StartSync(ctx context.Context, topics []string, onMessage OnKafkaMessageFunc) error {
	if err := eh.c.SubscribeTopics(topics, nil); err != nil {
		return fmt.Errorf("Failed to subscribe to topics: %s", err)
	}

	eh.loop(ctx, onMessage)
	return nil
}

func (eh *EventHandling) loop(ctx context.Context, onMessage OnKafkaMessageFunc) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down consumer goroutine")
			return
		default:
			msg, err := eh.c.ReadMessage(-1)
			if err == nil {
				var event PublicEvent
				if err := json.Unmarshal(msg.Value, &event); err != nil {
					log.Printf(`malformed message: %s`, err)
					continue
				}

				if err := onMessage(*msg.TopicPartition.Topic, event, msg.Value); err != nil {
					log.Printf(`on-message handler failed: %s`, err)
				}
				// TODO: We're not committing the offset.
			} else {
				// The client will automatically try to recover from all errors.
				log.Printf("Consumer error: %v (%v)\n", err, msg)
			}
		}
	}

	eh.c.Close()
}
