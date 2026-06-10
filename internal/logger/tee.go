package logger

import (
	"context"
	"log/slog"
)

// TeeHandler fans out each log record to multiple slog.Handlers.
type TeeHandler struct {
	handlers []slog.Handler
}

// NewTeeHandler returns a TeeHandler that writes to all provided handlers.
func NewTeeHandler(handlers ...slog.Handler) *TeeHandler {
	return &TeeHandler{handlers: handlers}
}

func (t *TeeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range t.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (t *TeeHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range t.handlers {
		if h.Enabled(ctx, r.Level) {
			_ = h.Handle(ctx, r.Clone())
		}
	}
	return nil
}

func (t *TeeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	updated := make([]slog.Handler, len(t.handlers))
	for i, h := range t.handlers {
		updated[i] = h.WithAttrs(attrs)
	}
	return &TeeHandler{handlers: updated}
}

func (t *TeeHandler) WithGroup(name string) slog.Handler {
	updated := make([]slog.Handler, len(t.handlers))
	for i, h := range t.handlers {
		updated[i] = h.WithGroup(name)
	}
	return &TeeHandler{handlers: updated}
}
