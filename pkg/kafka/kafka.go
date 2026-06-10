// Package kafka provides Kafka producer and consumer implementations for
// asynchronous message publishing and processing.
package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

// Producer writes messages to a Kafka topic.
type Producer struct {
	writer *kafka.Writer
}

// NewProducer creates a producer that publishes to the given topic.
func NewProducer(brokers []string, topic string) *Producer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		Balancer:     &kafka.LeastBytes{},
	}
	return &Producer{writer: w}
}

// Publish serializes and sends a message with the given key.
func (p *Producer) Publish(ctx context.Context, key, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(fmt.Sprintf("%v", key)),
		Value: data,
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	return nil
}

// Close shuts down the producer.
func (p *Producer) Close() error {
	return p.writer.Close()
}

// Consumer reads messages from a Kafka topic using a consumer group.
type Consumer struct {
	reader *kafka.Reader
}

// NewConsumer creates a consumer in the given group.
func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	return &Consumer{reader: r}
}

// Consume blocks and invokes handler for each message until ctx is cancelled.
func (c *Consumer) Consume(ctx context.Context, handler func(ctx context.Context, key, value []byte) error) error {
	slog.Info("consumer starting", "topic", c.reader.Config().Topic)

	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			slog.Error("read message error", "error", err, "topic", c.reader.Config().Topic)
			continue
		}

		if err := handler(ctx, msg.Key, msg.Value); err != nil {
			slog.Error("handle message error", "error", err, "topic", c.reader.Config().Topic)
		}
	}
}

// Close shuts down the consumer.
func (c *Consumer) Close() error {
	return c.reader.Close()
}

// EnsureTopic creates the topic if it does not already exist.
func EnsureTopic(brokers []string, topic string, partitions int) error {
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("dial kafka: %w", err)
	}
	defer func() { _ = conn.Close() }()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     partitions,
			ReplicationFactor: 1,
		},
	}

	if err := conn.CreateTopics(topicConfigs...); err != nil {
		return fmt.Errorf("create topic: %w", err)
	}

	slog.Info("kafka topic ensured", "topic", topic)
	return nil
}
