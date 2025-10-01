package network

import (
	"context"
	"errors"
	"log"

	"github.com/vaastav/columbo_go/components"
	"github.com/vaastav/columbo_go/events"
	"github.com/vaastav/columbo_go/trace"
)

var cntr int

type NetworkTraceGen struct {
	*components.BasePlugin
	InStream components.Plugin
	pending  map[string]*trace.ColumboTrace
}

func NewNetworkTraceGen(ctx context.Context, ins components.Plugin, buffer_size int, ID int) (*NetworkTraceGen, error) {
	outs, err := components.NewDataStream(ctx, make(chan *trace.ColumboTrace, buffer_size))
	if err != nil {
		return nil, err
	}

	nt := &NetworkTraceGen{
		components.NewBasePlugin(ID, outs),
		ins,
		make(map[string]*trace.ColumboTrace),
	}

	return nt, nil
}

func mergeTraces(t1, t2 *trace.ColumboTrace) *trace.ColumboTrace {
	t := &trace.ColumboTrace{}
	t.Graph = make(map[string][]string)
	t.Attributes = make(map[string]string)
	t.Type = trace.SPAN
	t.Attributes["span_type"] = "switch"

	net_span := trace.MergeSpans(t1.Spans[0], t1.Spans[0])
	t.Spans = append(t.Spans, net_span)
	t.Graph[net_span.ID] = []string{}
	return t
}

func (n *NetworkTraceGen) processTrace(t *trace.ColumboTrace) {
	// Just push the incomin span and trace types
	if t.Type == trace.SPAN || t.Type == trace.TRACE {
		n.OutStream.Push(t)
		return
	}
	// Check if this is an enqueue event
	event_type := t.Attributes["event_type"]
	if event_type == events.KNetworKEnqueueT.String() {
		transient_id := t.Attributes["transient_id"]
		n.pending[transient_id] = t
	}
	// Check if this is a dequeue or drop event
	if event_type == events.KNetworKDequeueT.String() || event_type == events.KNetworKDropT.String() {
		transient_id := t.Attributes["transient_id"]
		if v, ok := n.pending[transient_id]; ok {
			merged_trace := mergeTraces(v, t)
			n.OutStream.Push(merged_trace)
		} else {
			log.Println("Got dequeue/drop event for an unavailable transient id ")
		}
	}
}

func (n *NetworkTraceGen) Run(ctx context.Context) error {
	ins := n.InStream.GetOutDataStream()
	if ins == nil {
		return errors.New("Outdatastream of incoming plugin is nil")
	}
	for {
		select {
		case t, ok := <-ins.Data:
			if !ok {
				// Channel is closed so we can be done too
				n.OutStream.Close()
				return nil
			}
			n.processTrace(t)
		case <-ctx.Done():
			log.Println("Context is done. Quitting.")
			n.OutStream.Close()
			return nil
		}
	}
}

func (n *NetworkTraceGen) IncomingPlugins() []components.Plugin {
	return []components.Plugin{n.InStream}
}
