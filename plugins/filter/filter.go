package filter

import (
	"context"
	"log"

	"github.com/vaastav/columbo_go/components"
	"github.com/vaastav/columbo_go/trace"
)

type Filter struct {
	Op        func(t *trace.ColumboTrace) bool
	InStream  *components.DataStream
	OutStream *components.DataStream
}

func NewFilter(ctx context.Context, Op func(t *trace.ColumboTrace) bool, ins *components.DataStream, buffer_size int) (*Filter, error) {
	outs, err := components.NewDataStream(ctx, make(chan *trace.ColumboTrace, buffer_size))
	if err != nil {
		return nil, err
	}

	f := &Filter{
		Op:        Op,
		InStream:  ins,
		OutStream: outs,
	}

	return f, nil
}

func (f *Filter) Run(ctx context.Context) error {
	for {
		select {
		case t := <-f.InStream.Data:
			if f.Op(t) {
				f.OutStream.Push(t)
			}
		case <-ctx.Done():
			log.Println("Context is done. Quitting")
			f.OutStream.Close()
			return nil
		}
	}
}
