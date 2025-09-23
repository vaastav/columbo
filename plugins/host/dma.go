package host

import (
	"context"
	"log"

	"github.com/vaastav/columbo_go/components"
	"github.com/vaastav/columbo_go/events"
	"github.com/vaastav/columbo_go/trace"
)

type HostDmaTraceGen struct {
	InStream       *components.DataStream
	OutStream      *components.DataStream
	pending_reads  map[string]*trace.ColumboTrace
	pending_writes map[string]*trace.ColumboTrace
}

func NewHostDmaTraceGen(ctx context.Context, ins *components.DataStream, buffer_size int) (*HostDmaTraceGen, error) {
	outs, err := components.NewDataStream(ctx, make(chan *trace.ColumboTrace, buffer_size))
	if err != nil {
		return nil, err
	}

	tg := &HostDmaTraceGen{
		InStream:       ins,
		OutStream:      outs,
		pending_reads:  make(map[string]*trace.ColumboTrace),
		pending_writes: make(map[string]*trace.ColumboTrace),
	}

	return tg, nil
}

func (p *HostDmaTraceGen) processTrace(t *trace.ColumboTrace) {
	// Push any incoming span and trace types
	if t.Type == trace.SPAN || t.Type == trace.TRACE {
		p.OutStream.Push(t)
		return
	}
	event_type := t.Attributes["event_type"]
	if event_type == events.KHostDmaRT.String() {
		id := t.Attributes["id"]
		p.pending_reads[id] = t
	} else if event_type == events.KHostDmaWT.String() {
		// Add a new pending write
		id := t.Attributes["id"]
		p.pending_writes[id] = t
	} else if event_type == events.KHostDmaCT.String() {
		id := t.Attributes["id"]
		if v, ok := p.pending_reads[id]; ok {
			new_t := mergeTraces(v, t, "host dma read")
			p.OutStream.Push(new_t)
		} else if v, ok2 := p.pending_writes[id]; ok2 {
			new_t := mergeTraces(v, t, "host dma write")
			p.OutStream.Push(new_t)
		} else {
			log.Println("Completion for a dma id that was never issued. Skipping.")
		}
	} else {
		p.OutStream.Push(t)
	}
}
