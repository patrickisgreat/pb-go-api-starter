// Package logger configures log/slog for the service and hands out loggers
// that carry trace context.
//
// Init is called once at start-up. Get(ctx) returns a logger for any function
// that has a context; when the context carries an OpenTelemetry span the
// returned logger adds trace_id and span_id so log lines correlate with traces.
package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/trace"
)

// Options controls how the process logger is built.
type Options struct {
	// Level is debug, info, warn or error. Unknown values fall back to info.
	Level string
	// Pretty switches from JSON to human readable text output.
	Pretty bool
	// Service is added to every record as the "service" attribute and names
	// the OpenTelemetry log scope.
	Service string
	// ExportToOTel also sends records to the global OpenTelemetry logger
	// provider, so they reach the same backend as traces and metrics.
	ExportToOTel bool
}

type ctxKey struct{}

// Init builds the process logger from opts, installs it as slog's default and
// returns it.
func Init(opts Options) *slog.Logger {
	handlerOpts := &slog.HandlerOptions{Level: ParseLevel(opts.Level)}

	var handler slog.Handler
	if opts.Pretty {
		handler = slog.NewTextHandler(os.Stdout, handlerOpts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, handlerOpts)
	}
	if opts.ExportToOTel {
		handler = newFanoutHandler(handler, otelslog.NewHandler(opts.Service))
	}

	log := slog.New(handler)
	if opts.Service != "" {
		log = log.With(slog.String("service", opts.Service))
	}
	slog.SetDefault(log)
	return log
}

// ParseLevel converts a level name to a slog.Level, defaulting to info.
func ParseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithContext returns a child context that carries log. Get(ctx) will return
// it in preference to the default logger.
func WithContext(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, log)
}

// Get returns the logger stored in ctx, or the default logger, decorated with
// the trace and span ids of the current span when there is one.
func Get(ctx context.Context) *slog.Logger {
	log, ok := ctx.Value(ctxKey{}).(*slog.Logger)
	if !ok || log == nil {
		log = slog.Default()
	}
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		log = log.With(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}
	return log
}
