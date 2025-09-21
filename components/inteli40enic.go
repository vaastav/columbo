package components

import (
	"context"

	"github.com/vaastav/columbo_go/events"
	"github.com/vaastav/columbo_go/trace"
)

type NIC struct {
	*baseComponent
}

func NewNIC(ctx context.Context, Name string, ID int, buffer_size int) (*NIC, error) {
	t, err := trace.NewColumboTracer(Name)
	if err != nil {
		return nil, err
	}
	outs, err := NewDataStream(ctx, make(chan *trace.ColumboTrace, buffer_size))
	if err != nil {
		return nil, err
	}
	NIC_comp := &NIC{
		&baseComponent{
			Tracer:    t,
			Name:      Name,
			ID:        ID,
			OutStream: outs,
		},
	}
	return NIC_comp, nil
}

func (c *NIC) HandleEvent(event events.Event) error {
	// Filter out any event that is not a NIC event
	if event.Type >= events.KNicMsixT && event.Type <= events.KNicRxT {
		ct := traceFromEvent(c.Tracer, event)
		c.OutStream.Push(ct)
	}
	return nil
}
