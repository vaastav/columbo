package components

import (
	"github.com/vaastav/columbo_go/events"
	"github.com/vaastav/columbo_go/trace"
)

func spanFromEvent(tracer *trace.ColumboTracer, event events.Event) trace.ColumboSpan {
	cs := trace.ColumboSpan{
		BaseTracer: tracer,
		Name:       event.Type.String(),
		ID:         event.ParserName + "_" + event.ID,
	}
	cs.Events = append(cs.Events, event)
	return cs
}

func traceFromEvent(tracer *trace.ColumboTracer, event events.Event) *trace.ColumboTrace {
	cs := spanFromEvent(tracer, event)
	ct := &trace.ColumboTrace{Graph: make(map[string][]string)}
	ct.Spans = append(ct.Spans, cs)
	ct.Graph[cs.ID] = []string{}
	return ct
}
