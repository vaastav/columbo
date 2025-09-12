package trace

type ColumboTrace struct {
	Spans []ColumboSpan
	Graph map[string][]string // Save the connections between the spans
}
