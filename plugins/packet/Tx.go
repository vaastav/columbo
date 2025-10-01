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
	t1.Spans[0].Events = append(t1.Spans[0].Events, t2.Spans[0].Events...)
	t1.Type = trace.SPAN
	t1.Attributes["span_type"] = "host tx"
}

func mergeHostNicTrace(host_trace, nic_trace *trace.ColumboTrace) *trace.ColumboTrace {
	new_t := &trace.ColumboTrace{}
	new_t.Spans = append(new_t.Spans, host_trace.Spans...)
	new_t.Spans = append(new_t.Spans, nic_trace.Spans...)
	new_t.Attributes = make(map[string]string)
	new_t.Type = trace.TRACE
	new_t.Graph = make(map[string][]string)
	for k, v := range host_trace.Graph {
		new_t.Graph[k] = v
	}
	for k, v := range nic_trace.Graph {
		new_t.Graph[k] = v
	}
	if v, ok := new_t.Graph[host_trace.Spans[0].ID]; !ok {
		new_t.Graph[host_trace.Spans[0].ID] = []string{nic_trace.Spans[0].ID}
	} else {
		new_t.Graph[host_trace.Spans[0].ID] = append(v, nic_trace.Spans[0].ID)
	}
	new_t.Attributes["trace_type"] = "host+nic tx"

	return new_t
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
		new_t := mergeHostNicTrace(p.hostTrace, p.nicTrace)
		p.OutStream.Push(new_t)
		p.hostTrace = nil
		p.nicTrace = nil
	} else {
		p.OutStream.Push(t)
	}
}

func (p *PacketTx) processNicTrace(t *trace.ColumboTrace) {
	if t.Type == trace.EVENT || t.Type == trace.SPAN {
		p.OutStream.Push(t)
	}
	if t.Type == trace.TRACE {
		trace_type := t.Attributes["trace_type"]
		if trace_type == "nic tx" {
			p.nicTrace = t
		} else {
			p.OutStream.Push(t)
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
