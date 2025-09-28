package packet

import (
	"context"
	"errors"

	"github.com/vaastav/columbo_go/components"
	"github.com/vaastav/columbo_go/events"
	"github.com/vaastav/columbo_go/trace"
)

type PacketRx struct {
	*components.BasePlugin
	HostStream components.Plugin
	NICStream  components.Plugin
	hostTrace  *trace.ColumboTrace
	nicTrace   *trace.ColumboTrace
	msixTrace  *trace.ColumboTrace
}

func NewPacketRx(ctx context.Context, host_stream components.Plugin, nic_stream components.Plugin, buffer_size int, ID int) (*PacketRx, error) {
	outs, err := components.NewDataStream(ctx, make(chan *trace.ColumboTrace, buffer_size))
	if err != nil {
		return nil, err
	}
	pkt_rx := &PacketRx{
		components.NewBasePlugin(ID, outs),
		host_stream,
		nic_stream,
		nil,
		nil,
		nil,
	}
	return pkt_rx, nil
}

func (p *PacketRx) mergeHostNicTrace() *trace.ColumboTrace {
	new_t := &trace.ColumboTrace{}
	new_t.Spans = append(new_t.Spans, p.hostTrace.Spans...)
	new_t.Spans = append(new_t.Spans, p.nicTrace.Spans...)
	new_t.Attributes = make(map[string]string)
	new_t.Type = trace.TRACE
	new_t.Graph = make(map[string][]string)
	for k, v := range p.hostTrace.Graph {
		new_t.Graph[k] = v
	}
	for k, v := range p.nicTrace.Graph {
		new_t.Graph[k] = v
	}
	if v, ok := new_t.Graph[p.hostTrace.Spans[0].ID]; !ok {
		new_t.Graph[p.hostTrace.Spans[0].ID] = []string{p.nicTrace.Spans[0].ID}
	} else {
		new_t.Graph[p.hostTrace.Spans[0].ID] = append(v, p.nicTrace.Spans[0].ID)
	}
	new_t.Attributes["trace_type"] = "host+nic tx"

	return new_t
}

func (p *PacketRx) processHostTrace(t *trace.ColumboTrace) {
	if t.Type == trace.EVENT || t.Type == trace.TRACE {
		p.OutStream.Push(t)
	}
	span_type := t.Attributes["span_type"]
	if span_type == "host dma write" {
		new_t := p.mergeHostNicTrace()
		p.OutStream.Push(new_t)
		p.hostTrace = nil
		p.nicTrace = nil
	}
}

func (p *PacketRx) processNicTrace(t *trace.ColumboTrace) {
	if t.Type == trace.SPAN {
		p.OutStream.Push(t)
	}
	if t.Type == trace.EVENT {
		event_type := t.Attributes["event_type"]
		if event_type == events.KNicMsixT.String() {
			p.msixTrace = t
		} else {
			p.OutStream.Push(t)
		}
	}
	if t.Type == trace.TRACE {
		trace_type := t.Attributes["trace_type"]
		if trace_type == "nic rx" {
			p.nicTrace = t
		} else {
			p.OutStream.Push(t)
		}
	}
}

func (p *PacketRx) Run(ctx context.Context) error {
	hins := p.HostStream.GetOutDataStream()
	if hins == nil {
		return errors.New("host datastream of incoming plugin is nil")
	}
	nins := p.NICStream.GetOutDataStream()
	if nins == nil {
		return errors.New("nic datastream of incoming plugin is nil")
	}
	for {
		select {
		case t := <-hins.Data:
			p.processHostTrace(t)
		case t := <-nins.Data:
			p.processNicTrace(t)
		case <-ctx.Done():
			return p.Shutdown(ctx)
		}
	}
}

func (p *PacketRx) IncomingPlugins() []components.Plugin {
	return []components.Plugin{p.HostStream, p.NICStream}
}
