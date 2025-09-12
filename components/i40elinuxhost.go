package components

import (
	"context"

	"github.com/vaastav/columbo_go/events"
	"github.com/vaastav/columbo_go/trace"
)

type Host struct {
	*baseComponent
}

func NewHost(ctx context.Context, Name string, ID int, buffer_size int) (*Host, error) {
	t := trace.NewColumboTracer(Name)
	outs, err := NewDataStream(ctx, make(chan *trace.ColumboTrace, buffer_size))
	if err != nil {
		return nil, err
	}
	host_comp := &Host{
		&baseComponent{
			Tracer:    t,
			Name:      Name,
			ID:        ID,
			OutStream: outs,
		},
	}
	return host_comp, nil
}

func (c *Host) HandleEvent(event events.Event) error {
	// Filter out any event that is not a host event
	if event.Type >= events.KHostInstrT && event.Type <= events.KHostPciRWT {
		ct := traceFromEvent(c.Tracer, event)
		c.OutStream.Push(ct)
	}
	return nil
}
