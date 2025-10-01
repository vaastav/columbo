package host

import "github.com/vaastav/columbo_go/trace"

func mergeTraces(t1, t2 *trace.ColumboTrace, span_type string) *trace.ColumboTrace {
	t := &trace.ColumboTrace{}
	t.Graph = make(map[string][]string)
	t.Attributes = make(map[string]string)
	t.Type = trace.SPAN
	t.Attributes["span_type"] = span_type

	span := trace.MergeSpans(t1.Spans[0], t2.Spans[0])
	span.Name = span_type
	t.Spans = append(t.Spans, span)
	t.Graph[span.ID] = []string{}
	return t
}
