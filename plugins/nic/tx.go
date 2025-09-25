package nic

import (
	"context"
	"log"

	"github.com/vaastav/columbo_go/components"
	"github.com/vaastav/columbo_go/events"
	"github.com/vaastav/columbo_go/trace"
)

type NicTx struct {
	InStream   *components.DataStream
	OutStream  *components.DataStream
	pending_tx *trace.ColumboTrace
}

func NewNicTx(ctx context.Context, ins *components.DataStream, buffer_size int) (*NicTx, error) {
	outs, err := components.NewDataStream(ctx, make(chan *trace.ColumboTrace, buffer_size))
	if err != nil {
		return nil, err
	}

	tx := &NicTx{
		InStream:   ins,
		OutStream:  outs,
		pending_tx: nil,
	}
	return tx, nil
}

func addEthTxEvent(t, eth_tx *trace.ColumboTrace) {
	// Add eth tx span to root span
	t.Spans[0].Events = append(t.Spans[0].Events, eth_tx.Spans[0].Events...)
	t.Type = trace.TRACE
	t.Attributes["trace_type"] = "nic tx"
}

func addDMASpanToTxTrace(t, dma *trace.ColumboTrace) {
	t.Spans = append(t.Spans, dma.Spans...)
	t.Graph[t.Spans[0].ID] = append(t.Graph[t.Spans[0].ID], dma.Spans[0].ID)
	t.Graph[dma.Spans[0].ID] = []string{}
}

func (p *NicTx) processTrace(t *trace.ColumboTrace) {
	if t.Type == trace.TRACE {
		p.OutStream.Push(t)
		return
	}
	if t.Type == trace.SPAN {
		span_type := t.Attributes["span_type"]
		if span_type == "Nic DMA Read" {
			if p.pending_tx == nil {
				p.pending_tx = t
			} else {
				addDMASpanToTxTrace(p.pending_tx, t)
			}
		} else {
			// Push any other kind of spans
			p.OutStream.Push(t)
		}
	}
	if t.Type == trace.EVENT {
		event_type := t.Attributes["event_type"]
		if event_type == events.KNicTxT.String() {
			addEthTxEvent(p.pending_tx, t)
			// We can now push the trace forward
			p.OutStream.Push(p.pending_tx)
			p.pending_tx = nil
		} else if event_type == events.KNicMmioRT.String() {
			if p.pending_tx != nil {
				log.Println("Overwriting the previous pending tx span")
			}
			p.pending_tx = t
		}
	}
}

func (p *NicTx) Run(ctx context.Context) error {
	for {
		select {
		case t := <-p.InStream.Data:
			p.processTrace(t)
		case <-ctx.Done():
			p.OutStream.Close()
			return nil
		}
	}
}
