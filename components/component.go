package components

import (
	"context"
	"errors"

	"github.com/vaastav/columbo_go/events"
	"github.com/vaastav/columbo_go/trace"
)

type Component interface {
	GetTracer() *trace.ColumboTracer
	GetOutDataStream() *DataStream
	HandleEvent(event events.Event) error
	Shutdown(ctx context.Context) error
}

type baseComponent struct {
	Tracer    *trace.ColumboTracer
	Name      string
	ID        int
	OutStream *DataStream
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
