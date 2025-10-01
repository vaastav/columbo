package host

import (
	"context"
	"errors"
	"log"

	"github.com/vaastav/columbo_go/components"
	"github.com/vaastav/columbo_go/events"
	"github.com/vaastav/columbo_go/trace"
)

type HostMmioTraceGen struct {
	*components.BasePlugin
	InStream              components.Plugin
	pending_reads         map[string]*trace.ColumboTrace
	pending_writes        map[string]*trace.ColumboTrace
	pending_posted_writes map[string]*trace.ColumboTrace
}

func NewHostMmioTraceGen(ctx context.Context, ins components.Plugin, buffer_size int, ID int) (*HostMmioTraceGen, error) {
	outs, err := components.NewDataStream(ctx, make(chan *trace.ColumboTrace, buffer_size))
	if err != nil {
		return nil, err
	}

	tg := &HostMmioTraceGen{
		components.NewBasePlugin(ID, outs),
		ins,
		make(map[string]*trace.ColumboTrace),
		make(map[string]*trace.ColumboTrace),
		make(map[string]*trace.ColumboTrace),
	}

	return tg, nil
}

func (p *HostMmioTraceGen) processTrace(t *trace.ColumboTrace) {
	// Push any incoming span and trace types
	if t.Type == trace.SPAN || t.Type == trace.TRACE {
		p.OutStream.Push(t)
		return
	}
	event_type := t.Attributes["event_type"]
	if event_type == events.KHostMmioRT.String() {
		// Add a new pending read
		id := t.Attributes["id"]
		p.pending_reads[id] = t
	} else if event_type == events.KHostMmioWT.String() {
		// Add a new pending write
		is_posted := (t.Attributes["posted"] == "1")
		id := t.Attributes["id"]
		if is_posted {
			p.pending_posted_writes[id] = t
		} else {
			p.pending_writes[id] = t
		}
	} else if event_type == events.KHostMmioCRT.String() {
		// Complete the pending read
		id := t.Attributes["id"]
		if v, ok := p.pending_reads[id]; !ok {
			log.Println("No available pending read!")
		} else {
			new_t := mergeTraces(v, t, "host mmio read")
			p.OutStream.Push(new_t)
		}
	} else if event_type == events.KHostMmioCWT.String() {
		// Complete the pending write
		id := t.Attributes["id"]
		if v, ok := p.pending_writes[id]; !ok {
			log.Println("No available pending write!")
		} else {
			new_t := mergeTraces(v, t, "host mmio write")
			p.OutStream.Push(new_t)
		}
	} else if event_type == events.KHostMmioImRespPoWT.String() {
		// Complete the posted write
		id := t.Attributes["id"]
		if v, ok := p.pending_posted_writes[id]; !ok {
			log.Println("No available posted write!")
		} else {
			new_t := mergeTraces(v, t, "host mmio posted write")
			p.OutStream.Push(new_t)
		}
	} else {
		p.OutStream.Push(t)
	}
}

func (p *HostMmioTraceGen) Run(ctx context.Context) error {
	ins := p.InStream.GetOutDataStream()
	if ins == nil {
		return errors.New("Outdatastream of incoming plugin is nil")
	}
	for {
		select {
		case t := <-ins.Data:
			p.processTrace(t)
		case <-ctx.Done():
			log.Println("Context is done. Quitting.")
			p.OutStream.Close()
			return nil
		}
	}
}

func (p *HostMmioTraceGen) IncomingPlugins() []components.Plugin {
	return []components.Plugin{p.InStream}
}
