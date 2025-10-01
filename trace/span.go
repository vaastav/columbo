package trace

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/vaastav/columbo_go/events"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var SCALE_UNIT = time.Nanosecond // Maybe change to time.Microsecond?

type ColumboSpan struct {
	BaseTracer *ColumboTracer
	Name       string
	ID         string
	Events     []events.Event
}

func MergeSpans(s1, s2 ColumboSpan) ColumboSpan {
	res := ColumboSpan{}
	res.BaseTracer = s1.BaseTracer
	res.Events = append(res.Events, s1.Events...)
	res.Events = append(res.Events, s2.Events...)
	res.ID = uuid.New().String()
	return res
}

func (c *ColumboSpan) ExportSpan(ctx context.Context, start_time time.Time, opts ...trace.SpanStartOption) context.Context {
	// Sort all the the events in the span
	sort.Slice(c.Events, func(i, j int) bool {
		return c.Events[i].Timestamp < c.Events[j].Timestamp
	})

	var span trace.Span
	for idx, e := range c.Events {
		ts := start_time.Add(time.Duration(e.Timestamp) * SCALE_UNIT)
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
	// Return the updated context so we can use that for child spans
	return ctx
}
