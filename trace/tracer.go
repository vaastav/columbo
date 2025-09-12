package trace

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type ColumboTracer struct {
	base trace.Tracer
}

func NewColumboTracer(name string) *ColumboTracer {
	return &ColumboTracer{otel.Tracer(name)}
}

func (ct *ColumboTracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return ct.base.Start(ctx, name, opts...)
}

func (ct *ColumboTracer) tracer() {}
