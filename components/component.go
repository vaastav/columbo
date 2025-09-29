package components

import (
	"context"
	"errors"
	"sync"

	"github.com/vaastav/columbo_go/events"
	"github.com/vaastav/columbo_go/trace"
)

type Component interface {
	Plugin
	GetTracer() *trace.ColumboTracer
	HandleEvent(event events.Event) error
	Stop()
}

type baseComponent struct {
	*BasePlugin
	Tracer *trace.ColumboTracer
	Name   string
	stop   chan bool
}

func newBaseComponent(ID int, outs *DataStream, t *trace.ColumboTracer, Name string) *baseComponent {
	return &baseComponent{
		NewBasePlugin(ID, outs),
		t,
		Name,
		make(chan bool),
	}
}

func (c *baseComponent) GetTracer() *trace.ColumboTracer {
	return c.Tracer
}

func (c *baseComponent) GetOutDataStream() *DataStream {
	return c.OutStream
}

func (c *baseComponent) HandleEvent(event events.Event) error {
	return errors.New("Base component does not handle any event!")
}

func (c *baseComponent) Shutdown(ctx context.Context) error {
	// Close the outgoing stream
	c.OutStream.Close()
	// Shutdown the tracer
	return c.Tracer.Shutdown(ctx)
}

func (c *baseComponent) Launch(ctx context.Context, wg *sync.WaitGroup) {
	// Don't really need to do much here other than setting the Running flag
	c.Running = true
}

func (c *baseComponent) Stop() {
	c.stop <- true
}

func (c *baseComponent) Run(ctx context.Context) error {
	select {
	case <-c.stop:
		return nil
	case <-ctx.Done():
		return nil
	}
}
