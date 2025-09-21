package trace

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/trace"
)

type ColumboTrace struct {
	Spans      []ColumboSpan
	Graph      map[string][]string // Save the connections between the spans
	Attributes map[string]string
}

func topoSort(graph map[string][]string) ([]string, map[string][]string) {
	parents := make(map[string][]string)
	var result []string
	inDegree := make(map[string]int)
	for node, neighbors := range graph {
		if _, ok := inDegree[node]; !ok {
			inDegree[node] = 0
		}
		for _, neigh := range neighbors {
			if _, ok := inDegree[neigh]; !ok {
				inDegree[neigh] = 0
			}
			inDegree[neigh]++
		}

		if _, ok := parents[node]; !ok {
			parents[node] = []string{}
		}
		for _, neigh := range neighbors {
			if _, ok := parents[neigh]; !ok {
				parents[neigh] = []string{}
			}
			parents[neigh] = append(parents[neigh], node)
		}
	}

	queue := []string{}
	for node, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, node)
		}
	}

	for len(queue) > 0 {
		// Pop front
		n := queue[0]
		queue = queue[1:]
		result = append(result, n)

		// Reduce in-degree of neighbors
		for _, neigh := range graph[n] {
			inDegree[neigh]--
			if inDegree[neigh] == 0 {
				queue = append(queue, neigh)
			}
		}
	}

	return result, parents
}

func (ct *ColumboTrace) Export() {
	// Walk the graph and then individually export the spans (we probably need to do a TopoSort on the graph)
	// Ensure that the context is correctly propagated.
	// Ensure that for spans that have multiple parents we have the context of all the parents to embed them as links
	span_map := make(map[string]*ColumboSpan)
	for _, span := range ct.Spans {
		span_map[span.ID] = &span
	}

	span_order, parents_map := topoSort(ct.Graph)
	ts := time.Now()
	columbo_ctx := context.Background()
	columbo_ctx, root_span := root_tracer.Start(columbo_ctx, "root", trace.WithTimestamp(ts))
	defer root_span.End()
	contexts := make(map[string]context.Context)
	for _, span_id := range span_order {
		span := span_map[span_id]
		var ctx context.Context
		parents := parents_map[span_id]
		var opts []trace.SpanStartOption
		// If no parent then simply use background context
		if len(parents) == 0 {
			ctx = columbo_ctx
		} else {
			// Pick the context of the first parent
			ctx = contexts[parents[0]]
			var links []trace.Link
			for _, p := range parents {
				l := trace.LinkFromContext(contexts[p])
				links = append(links, l)
			}
			opts = append(opts, trace.WithLinks(links...))
		}

		// Export span
		ctx = span.ExportSpan(ctx, ts, opts...)

		contexts[span_id] = ctx
	}
}
