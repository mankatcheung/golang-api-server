package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/golang-api-server/internal/logger"
	"github.com/golang-api-server/pkg/kafka"
	"gorm.io/gorm"
)

// StartLogConsumer launches a background goroutine that reads from the app-logs
// Kafka topic and inserts each record into the PostgreSQL logs table.
func StartLogConsumer(brokers []string, topic, groupID string, db *gorm.DB) MessageConsumer {
	c := kafka.NewConsumer(brokers, topic, groupID)

	go func() {
		if err := c.Consume(context.Background(), makeLogHandler(db)); err != nil {
			slog.Error("log consumer stopped", "error", err)
		}
	}()

	return c
}

func makeLogHandler(db *gorm.DB) func(ctx context.Context, key, value []byte) error {
	return func(ctx context.Context, _ []byte, value []byte) error {
		var rec logger.LogRecord
		if err := json.Unmarshal(value, &rec); err != nil {
			return fmt.Errorf("unmarshal log record: %w", err)
		}

		attrsJSON, err := json.Marshal(rec.Attrs)
		if err != nil {
			return fmt.Errorf("marshal log attrs: %w", err)
		}

		return db.WithContext(ctx).Exec(
			`INSERT INTO logs (logged_at, level, message, attrs) VALUES (?, ?, ?, ?)`,
			rec.Time, rec.Level, rec.Message, attrsJSON,
		).Error
	}
}
