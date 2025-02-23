package kafkaproducer

import (
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// Producer struct
type Producer struct {
	Producer *kafka.Producer
	Topic    string
}

// NewProducer initializes a Kafka producer
func NewProducer(broker, topic string) (*Producer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": broker})
	if err != nil {
		return nil, err
	}

	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Delivery failed: %v\n", ev.TopicPartition.Error)
				} else {
					log.Printf("Message delivered to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	return &Producer{Producer: p, Topic: topic}, nil
}

// ProduceMessage sends a message to Kafka
func (p *Producer) ProduceMessage(key, value string) error {
	return p.Producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &p.Topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          []byte(value),
	}, nil)
}

// Close producer
func (p *Producer) Close() {
	p.Producer.Close()
}
