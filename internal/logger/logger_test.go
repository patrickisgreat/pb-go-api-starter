package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	qt "github.com/frankban/quicktest"
	"go.opentelemetry.io/otel/trace"
)

func TestParseLevel(t *testing.T) {
	c := qt.New(t)
	c.Assert(ParseLevel("debug"), qt.Equals, slog.LevelDebug)
	c.Assert(ParseLevel("INFO"), qt.Equals, slog.LevelInfo)
	c.Assert(ParseLevel("warn"), qt.Equals, slog.LevelWarn)
	c.Assert(ParseLevel("error"), qt.Equals, slog.LevelError)
	c.Assert(ParseLevel("nonsense"), qt.Equals, slog.LevelInfo)
}

func TestGetUsesContextLoggerAndTraceIDs(t *testing.T) {
	c := qt.New(t)

	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, nil))

	traceID, _ := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	spanID, _ := trace.SpanIDFromHex("0102030405060708")
	sc := trace.NewSpanContext(trace.SpanContextConfig{TraceID: traceID, SpanID: spanID})

	ctx := trace.ContextWithSpanContext(WithContext(context.Background(), log), sc)
	Get(ctx).Info("hello")

	var record map[string]any
	c.Assert(json.Unmarshal(buf.Bytes(), &record), qt.IsNil)
	c.Assert(record["msg"], qt.Equals, "hello")
	c.Assert(record["trace_id"], qt.Equals, traceID.String())
	c.Assert(record["span_id"], qt.Equals, spanID.String())
}

func TestFanoutReachesEveryHandler(t *testing.T) {
	c := qt.New(t)

	var a, b bytes.Buffer
	h := newFanoutHandler(slog.NewTextHandler(&a, nil), slog.NewTextHandler(&b, nil))
	slog.New(h).With("k", "v").Info("hi")

	c.Assert(a.String(), qt.Contains, "msg=hi")
	c.Assert(a.String(), qt.Contains, "k=v")
	c.Assert(b.String(), qt.Contains, "msg=hi")
}
