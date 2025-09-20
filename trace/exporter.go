package trace

import (
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

var root_tracer *ColumboTracer

var exp *jaeger.Exporter

// InitJaegerExporter initializes a jaeger exporter to send spans to the collector listening at the provided url
func InitJaegerExporter(url string) error {
	var err error
	exp, err = jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return err
	}

	root_tracer = NewColumboTracer("columbo")

	return nil
}

func NewTracerProvider(name string) *sdktrace.TracerProvider {
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceName(name))),
		sdktrace.WithBatcher(exp),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	return tp
}
