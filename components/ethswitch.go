package components

import (
	"context"

	"github.com/vaastav/columbo_go/events"
	"github.com/vaastav/columbo_go/trace"
)

type Switch struct {
	*baseComponent
}

func NewSwitch(ctx context.Context, Name string, ID int, buffer_size int) (*Switch, error) {
	t, err := trace.NewColumboTracer(Name)
	if err != nil {
		return nil, err
	}
	outs, err := NewDataStream(ctx, make(chan *trace.ColumboTrace, buffer_size))
	if err != nil {
		return nil, err
	}
	switch_comp := &Switch{
		newBaseComponent(ID, outs, t, Name),
	}
	return switch_comp, nil
}

func (c *Switch) HandleEvent(event events.Event) error {
	if event.Type == events.KNetworKEnqueueT || event.Type == events.KNetworKDequeueT || event.Type == events.KNetworKDropT {
		ct := traceFromEvent(c.Tracer, event)
		c.OutStream.Push(ct)
	}
	return nil
}
