// Package telemetry wires up the OpenTelemetry SDK and provides the HTTP and
// Twirp instrumentation used by cmd/app.
//
// Everything is configured from the standard OTEL_* environment variables:
// OTEL_SERVICE_NAME and OTEL_RESOURCE_ATTRIBUTES describe the service,
// OTEL_TRACES_EXPORTER, OTEL_METRICS_EXPORTER and OTEL_LOGS_EXPORTER choose
// exporters (otlp by default, "none" to switch one off), OTEL_TRACES_SAMPLER
// picks sampling. OTEL_SDK_DISABLED=true skips the whole setup, which is what
// the local and test environments do.
package telemetry

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/bakins/twirpotel"
	"github.com/twitchtv/twirp"
	"go.opentelemetry.io/contrib/detectors/aws/ecs"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Option customises Setup.
type Option func(*options)

type options struct {
	detectors []resource.Detector
}

// WithECSDetector adds the AWS ECS resource detector, so spans carry the
// cluster, task and container identity when running on ECS. It is harmless
// elsewhere: off ECS it detects nothing.
func WithECSDetector() Option {
	return func(o *options) {
		o.detectors = append(o.detectors, ecs.NewResourceDetector())
	}
}

// Disabled reports whether OTEL_SDK_DISABLED asks for no telemetry at all.
func Disabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("OTEL_SDK_DISABLED")))
	return v == "true" || v == "1"
}

// Setup installs global trace, metric and log providers and returns a function
// that flushes and stops them. When telemetry is disabled it installs nothing
// and the returned function is a no-op.
func Setup(ctx context.Context, opts ...Option) (func(context.Context) error, error) {
	if Disabled() {
		return func(context.Context) error { return nil }, nil
	}

	var o options
	for _, opt := range opts {
		opt(&o)
	}

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithDetectors(o.detectors...),
	)
	if err != nil {
		return nil, err
	}

	var shutdowns []func(context.Context) error
	shutdown := func(ctx context.Context) error {
		var errs []error
		for i := len(shutdowns) - 1; i >= 0; i-- {
			errs = append(errs, shutdowns[i](ctx))
		}
		return errors.Join(errs...)
	}
	fail := func(err error) (func(context.Context) error, error) {
		return nil, errors.Join(err, shutdown(ctx))
	}

	spanExporter, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return fail(err)
	}
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(spanExporter),
		sdktrace.WithResource(res),
	)
	shutdowns = append(shutdowns, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	metricReader, err := autoexport.NewMetricReader(ctx)
	if err != nil {
		return fail(err)
	}
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(metricReader),
		sdkmetric.WithResource(res),
	)
	shutdowns = append(shutdowns, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	logExporter, err := autoexport.NewLogExporter(ctx)
	if err != nil {
		return fail(err)
	}
	loggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
		sdklog.WithResource(res),
	)
	shutdowns = append(shutdowns, loggerProvider.Shutdown)
	otel.SetLoggerProvider(loggerProvider)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return shutdown, nil
}

// InstrumentHandler wraps h so that every request gets a server span and the
// standard HTTP request metrics, under the given operation name.
func InstrumentHandler(name string, h http.Handler) http.Handler {
	return otelhttp.NewHandler(h, name)
}

// WithTwirpServerInterceptor returns a Twirp server option that opens a span
// per RPC, named after the service and method.
func WithTwirpServerInterceptor() twirp.ServerOption {
	return twirp.WithServerInterceptors(twirpotel.ServerInterceptor())
}
