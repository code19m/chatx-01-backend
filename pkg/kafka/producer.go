package kafka

import (
	"chatx-01-backend/pkg/errs"
	"context"
	"fmt"
	"strings"

	"github.com/IBM/sarama"
)

// Message represents a Producer Kafka message with key, value, and headers.
type Message struct {
	Key     []byte
	Value   []byte
	Headers map[string]string
}

// Producer represents a Kafka producer.
type Producer struct {
	cfg          ProducerConfig
	topic        string
	serviceName  string
	saramaCfg    *sarama.Config
	syncProducer sarama.SyncProducer
}

// NewProducer creates a new Kafka producer.
func NewProducer(
	cfg ProducerConfig,
	topic string,
	serviceName string,
) (*Producer, error) {
	const op = "kafka.NewProducer"

	saramaCfg, err := cfg.getSaramaConfig(serviceName)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	// Create a new sync producer
	producer, err := sarama.NewSyncProducer(strings.Split(cfg.Brokers, ","), saramaCfg)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	return &Producer{
		cfg:          cfg,
		topic:        topic,
		serviceName:  serviceName,
		saramaCfg:    saramaCfg,
		syncProducer: producer,
	}, nil
}

// SendMessage sends a message to the configured Kafka topic.
func (p *Producer) SendMessage(ctx context.Context, m *Message) error {
	const op = "kafka.SendMessage"

	kafkaMsg := p.buildKafkaProducerMsg(ctx, m)

	// Produce message
	partition, offset, err := p.syncProducer.SendMessage(kafkaMsg)
	if err != nil {
		return errs.Wrap(
			op,
			fmt.Errorf(
				"send_error: %w, topic: %s, partition: %d, offset: %d",
				err, kafkaMsg.Topic, partition, offset,
			),
		)
	}

	return nil
}

// SendMessages sends multiple messages to the configured Kafka topic.
func (p *Producer) SendMessages(ctx context.Context, messages []Message) error {
	const op = "kafka.SendMessages"

	if len(messages) == 0 {
		return nil
	}

	kafkaMessages := make([]*sarama.ProducerMessage, len(messages))
	for i, m := range messages {
		kafkaMessages[i] = p.buildKafkaProducerMsg(ctx, &m)
	}

	err := p.syncProducer.SendMessages(kafkaMessages)
	if err != nil {
		return errs.Wrap(
			op,
			fmt.Errorf(
				"send_error: %w, topic: %s",
				err, kafkaMessages[0].Topic,
			),
		)
	}

	return nil
}

func (p *Producer) buildKafkaProducerMsg(ctx context.Context, m *Message) *sarama.ProducerMessage {
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.ByteEncoder(m.Key),
		Value: sarama.ByteEncoder(m.Value),
	}

	// Add headers to the message
	for k, v := range m.Headers {
		msg.Headers = append(msg.Headers, sarama.RecordHeader{
			Key:   []byte(k),
			Value: []byte(v),
		})
	}

	return msg
}

// Close closes the producer.
func (p *Producer) Close() error {
	const op = "kafka.Producer.Close"

	err := p.syncProducer.Close()
	if err != nil {
		return errs.Wrap(op, err)
	}

	return nil
}
