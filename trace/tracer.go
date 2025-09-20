package trace

import (
	"context"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type ColumboTracer struct {
	tp   *sdktrace.TracerProvider
	base trace.Tracer
}

func NewColumboTracer(name string) *ColumboTracer {
	tp := NewTracerProvider(name)
	tracer := tp.Tracer(name)
	return &ColumboTracer{tp, tracer}
}

func (ct *ColumboTracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return ct.base.Start(ctx, name, opts...)
}

func (ct *ColumboTracer) tracer() {}

func (ct *ColumboTracer) Shutdown(ctx context.Context) error {
	return ct.tp.Shutdown(ctx)
}
