package telemetry

import (
	"context"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestDisabled(t *testing.T) {
	c := qt.New(t)

	t.Setenv("OTEL_SDK_DISABLED", "true")
	c.Assert(Disabled(), qt.IsTrue)

	t.Setenv("OTEL_SDK_DISABLED", "false")
	c.Assert(Disabled(), qt.IsFalse)

	t.Setenv("OTEL_SDK_DISABLED", "")
	c.Assert(Disabled(), qt.IsFalse)
}

func TestSetupWhenDisabledIsNoop(t *testing.T) {
	c := qt.New(t)
	t.Setenv("OTEL_SDK_DISABLED", "true")

	shutdown, err := Setup(context.Background(), WithECSDetector())
	c.Assert(err, qt.IsNil)
	c.Assert(shutdown(context.Background()), qt.IsNil)
}

func TestSetupWithNoExporters(t *testing.T) {
	c := qt.New(t)
	t.Setenv("OTEL_SDK_DISABLED", "")
	t.Setenv("OTEL_TRACES_EXPORTER", "none")
	t.Setenv("OTEL_METRICS_EXPORTER", "none")
	t.Setenv("OTEL_LOGS_EXPORTER", "none")

	shutdown, err := Setup(context.Background())
	c.Assert(err, qt.IsNil)
	c.Assert(shutdown(context.Background()), qt.IsNil)
}
