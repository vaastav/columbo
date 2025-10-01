package filter

import (
	"context"
	"errors"
	"log"

	"github.com/vaastav/columbo_go/components"
	"github.com/vaastav/columbo_go/trace"
)

type Filter struct {
	*components.BasePlugin
	Op       func(t *trace.ColumboTrace) bool
	InStream components.Plugin
}

func NewFilter(ctx context.Context, ins components.Plugin, buffer_size int, ID int, Op func(t *trace.ColumboTrace) bool) (*Filter, error) {
	outs, err := components.NewDataStream(ctx, make(chan *trace.ColumboTrace, buffer_size))
	if err != nil {
		return nil, err
	}

	f := &Filter{
		components.NewBasePlugin(ID, outs),
		Op,
		ins,
	}
	f.OutStream = outs

	return f, nil
}

func (f *Filter) Run(ctx context.Context) error {
	ds := f.InStream.GetOutDataStream()
	if ds == nil {
		return errors.New("Outdatastream of incoming plugin is nil")
	}
	for {
		select {
		case t, ok := <-ds.Data:
			if !ok {
				// CHannel is closed and so are we
				f.OutStream.Close()
			}
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

func (f *Filter) IncomingPlugins() []components.Plugin {
	return []components.Plugin{f.InStream}
}
