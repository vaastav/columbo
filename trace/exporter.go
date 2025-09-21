package trace

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

var root_tracer *ColumboTracer

var URL string

type debugExporter struct {
	wrapped sdktrace.SpanExporter
}

func (d *debugExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	for _, s := range spans {
		fmt.Printf("Exporting span: name=%s traceID=%s spanID=%s start=%s end=%s scope=%s\n",
			s.Name(),
			s.SpanContext().TraceID().String(),
			s.SpanContext().SpanID().String(),
			s.StartTime().Format(time.RFC3339Nano),
			s.EndTime().Format(time.RFC3339Nano),
			s.InstrumentationScope().Name,
		)
	}
	return d.wrapped.ExportSpans(ctx, spans)
}

func (d *debugExporter) Shutdown(ctx context.Context) error {
	return d.wrapped.Shutdown(ctx)
}

// InitJaegerExporter initializes a jaeger exporter to send spans to the collector listening at the provided url
func InitJaegerExporter(url string) error {
	URL = url

	var err error
	root_tracer, err = NewColumboTracer("columbo")

	return err
}

func NewTracerProvider(name string) (*sdktrace.TracerProvider, error) {
	var err error
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(URL)))
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceName(name))),
		sdktrace.WithBatcher(exp),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	return tp, nil
}
