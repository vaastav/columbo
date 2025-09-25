package packet

import (
	"context"

	"github.com/vaastav/columbo_go/components"
	"github.com/vaastav/columbo_go/trace"
)

type PacketTx struct {
	HostStream *components.DataStream
	NICStream  *components.DataStream
	OutStream  *components.DataStream
	hostTrace  *trace.ColumboTrace
	nicTrace   *trace.ColumboTrace
}

func NewPacketTx(ctx context.Context, host_stream *components.DataStream, nic_stream *components.DataStream, buffer_size int) (*PacketTx, error) {
	outs, err := components.NewDataStream(ctx, make(chan *trace.ColumboTrace, buffer_size))
	if err != nil {
		return nil, err
	}

	pkt_tx := &PacketTx{
		HostStream: host_stream,
		NICStream:  nic_stream,
		OutStream:  outs,
	}
	return pkt_tx, nil
}

func mergeHostTxTrace(t1, t2 *trace.ColumboTrace) {
	mmio_span := t1.Spans[0].ID
	dma_span := t2.Spans[0].ID
	t1.Attributes["span_type"] = "host pkt tx"
	t1.Spans = append(t1.Spans, t2.Spans...)
	t1.Attributes["mmio"] = mmio_span
	t1.Attributes["dma"] = dma_span
}

func (p *PacketTx) processHostTrace(t *trace.ColumboTrace) {
	if t.Type == trace.EVENT || t.Type == trace.TRACE {
		p.OutStream.Push(t)
	}
	span_type := t.Attributes["span_type"]
	if span_type == "host mmio write" || span_type == "host mmio posted write" {
		// Expect to see a trace with span_type host mmio write
		p.hostTrace = t
	} else if span_type == "host dma read" {
		// Then we would get host dma
		mergeHostTxTrace(p.hostTrace, t)
	} else {
		p.OutStream.Push(t)
	}
}

func (p *PacketTx) processNicTrace(t *trace.ColumboTrace) {
	if t.Type == trace.TRACE || t.Type == trace.EVENT {
		p.OutStream.Push(t)
	}
	if t.Type == trace.SPAN {
		// We should get a nic mmio read
		span_type := t.Attributes["span_type"]
		if span_type == "nic tx" {
			p.nicTrace = t
		}
	}
}

func (p *PacketTx) Run(ctx context.Context) error {
	for {
		select {
		case t := <-p.HostStream.Data:
			p.processHostTrace(t)
		case t := <-p.NICStream.Data:
			p.processNicTrace(t)
		case <-ctx.Done():
			p.Shutdown()
			return nil
		}
	}
}

func (p *PacketTx) Shutdown() {
	p.OutStream.Close()
}
