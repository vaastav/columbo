package trace

type ColumboTrace struct {
	Spans      []ColumboSpan
	Graph      map[string][]string // Save the connections between the spans
	Attributes map[string]string
}

func (ct *ColumboTrace) Export() {
	// TODO: Implement this

	// Walk the graph and then individually export the spans (we probably need to do a TopoSort on the graph)
	// Ensure that the context is correctly propagated.
	// Ensure that for spans that have multiple parents we have the context of all the parents to embed them as links
}
