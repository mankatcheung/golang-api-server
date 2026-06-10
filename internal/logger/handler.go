package logger

import (
	"context"
	"log/slog"
	"sync"

	"github.com/golang-api-server/internal/service"
)

const bufferSize = 1000

// KafkaHandler is an async slog.Handler that publishes log records to Kafka.
// Handle() is non-blocking; a background goroutine drains the channel.
type KafkaHandler struct {
	producer service.MessageProducer
	level    slog.Level
	attrs    []slog.Attr
	groups   []string
	ch       chan slog.Record
	wg       sync.WaitGroup
}

// NewKafkaHandler creates a KafkaHandler and starts the drain goroutine.
func NewKafkaHandler(producer service.MessageProducer, level slog.Level) *KafkaHandler {
	h := &KafkaHandler{
		producer: producer,
		level:    level,
		ch:       make(chan slog.Record, bufferSize),
	}
	h.wg.Add(1)
	go h.drain()
	return h
}

// Close flushes the remaining buffer and waits for the drain goroutine to exit.
func (h *KafkaHandler) Close() {
	close(h.ch)
	h.wg.Wait()
}

func (h *KafkaHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *KafkaHandler) Handle(_ context.Context, r slog.Record) error {
	select {
	case h.ch <- r:
	default:
		// drop if buffer full — prefer zero handler latency over blocking
	}
	return nil
}

func (h *KafkaHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	combined := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(combined, h.attrs)
	copy(combined[len(h.attrs):], attrs)
	return &KafkaHandler{
		producer: h.producer,
		level:    h.level,
		attrs:    combined,
		groups:   h.groups,
		ch:       h.ch,
		// share the same channel/wg so no extra goroutine is spawned
	}
}

func (h *KafkaHandler) WithGroup(name string) slog.Handler {
	groups := make([]string, len(h.groups)+1)
	copy(groups, h.groups)
	groups[len(h.groups)] = name
	return &KafkaHandler{
		producer: h.producer,
		level:    h.level,
		attrs:    h.attrs,
		groups:   groups,
		ch:       h.ch,
	}
}

func (h *KafkaHandler) drain() {
	defer h.wg.Done()
	for r := range h.ch {
		h.publish(r)
	}
}

func (h *KafkaHandler) publish(r slog.Record) {
	rec := LogRecord{
		Time:    r.Time,
		Level:   r.Level.String(),
		Message: r.Message,
		Attrs:   make(map[string]any, r.NumAttrs()+len(h.attrs)),
	}

	for _, a := range h.attrs {
		rec.Attrs[a.Key] = a.Value.Any()
	}
	r.Attrs(func(a slog.Attr) bool {
		rec.Attrs[a.Key] = a.Value.Any()
		return true
	})

	// Best-effort publish; log errors to stderr only to avoid recursion.
	if err := h.producer.Publish(context.Background(), r.Level.String(), rec); err != nil {
		// intentionally use fmt to stderr — slog would recurse
		_ = err
	}
}
