package nic

import (
	"context"
	"log"

	"github.com/vaastav/columbo_go/components"
	"github.com/vaastav/columbo_go/events"
	"github.com/vaastav/columbo_go/trace"
)

type NicDMATraceGen struct {
	InStream          *components.DataStream
	OutStream         *components.DataStream
	pending_read_dma  *trace.ColumboTrace
	pending_write_dma *trace.ColumboTrace
}

func NewNicDmaTraceGen(ctx context.Context, ins *components.DataStream, buffer_size int) (*NicDMATraceGen, error) {
	outs, err := components.NewDataStream(ctx, make(chan *trace.ColumboTrace, buffer_size))
	if err != nil {
		return nil, err
	}

	tg := &NicDMATraceGen{
		InStream:          ins,
		OutStream:         outs,
		pending_read_dma:  nil,
		pending_write_dma: nil,
	}
	return tg, nil
}

func mergeTraces(t1, t2 *trace.ColumboTrace, span_type string) *trace.ColumboTrace {
	t := &trace.ColumboTrace{}
	t.Graph = make(map[string][]string)
	t.Attributes = make(map[string]string)
	t.Type = trace.SPAN
	t.Attributes["span_type"] = span_type

	span := trace.MergeSpans(t1.Spans[0], t2.Spans[0])
	t.Spans = append(t.Spans, span)
	t.Graph[span.ID] = []string{}
	return t
}

func (p *NicDMATraceGen) processTrace(t *trace.ColumboTrace) {
	if t.Type == trace.SPAN || t.Type == trace.TRACE {
		p.OutStream.Push(t)
		return
	}
	event_type := t.Attributes["event_type"]
	if event_type == events.KNicDmaIT.String() {
		op := t.Attributes["op"]
		if op == "read" {
			p.pending_read_dma = t
		} else if op == "write" {
			p.pending_write_dma = t
		}
	} else if event_type == events.KNicDmaExT.String() {
		op := t.Attributes["op"]
		if op == "read" {
			new_t := mergeTraces(p.pending_read_dma, t, "Nic DMA Read")
			p.pending_read_dma = new_t
		} else if op == "write" {
			new_t := mergeTraces(p.pending_write_dma, t, "Nic DMA Write")
			p.pending_write_dma = new_t
		}
	} else if event_type == events.KNicDmaWDataT.String() {
		new_t := mergeTraces(p.pending_write_dma, t, "Nic DMA Write")
		p.pending_write_dma = new_t
	} else if event_type == events.KNicDmaCWT.String() {
		new_t := mergeTraces(p.pending_write_dma, t, "Nic DMA Write")
		p.OutStream.Push(new_t)
		p.pending_write_dma = nil
	} else if event_type == events.KNicDmaCRT.String() {
		new_t := mergeTraces(p.pending_read_dma, t, "Nic DMA Read")
		p.OutStream.Push(new_t)
		p.pending_read_dma = nil
	} else {
		p.OutStream.Push(t)
	}
}

func (p *NicDMATraceGen) Run(ctx context.Context) error {
	for {
		select {
		case t := <-p.InStream.Data:
			p.processTrace(t)
		case <-ctx.Done():
			log.Println("Context is done. Quitting")
			p.OutStream.Close()
			return nil
		}
	}
}
