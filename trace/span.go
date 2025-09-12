package trace

import (
	"context"
	"sort"
	"time"

	"github.com/vaastav/columbo_go/events"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ColumboSpan struct {
	BaseTracer *ColumboTracer
	Name       string
	ID         string
	Events     []events.Event
}

func (c *ColumboSpan) ExportSpan(ctx context.Context, start_time uint64, opts ...trace.SpanStartOption) {
	// Sort all the the events in the span
	sort.Slice(c.Events, func(i, j int) bool {
		return c.Events[i].Timestamp < c.Events[j].Timestamp
	})

	var span trace.Span
	for idx, e := range c.Events {
		ts := time.Unix(0, int64(start_time+e.Timestamp))
		if idx == 0 {
			// Start the span
			ctx, span = c.BaseTracer.Start(ctx, c.Name, append(opts, trace.WithTimestamp(ts))...)
		}

		attributes := []attribute.KeyValue{}
		for k, v := range e.Attributes {
			attributes = append(attributes, attribute.String(k, v))
		}
		attributes = append(attributes, attribute.String("parser", e.ParserName))
		attributes = append(attributes, attribute.String("parser", e.Type.String()))
		span.AddEvent(e.Type.String()+e.Message, trace.WithAttributes(attributes...), trace.WithTimestamp(ts))
		if idx == len(c.Events)-1 {
			// End the span
			span.End(trace.WithTimestamp(ts))
		}
	}
}
