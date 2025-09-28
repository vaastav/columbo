package discard

import (
	"context"
	"log"

	"github.com/vaastav/columbo_go/components"
	"github.com/vaastav/columbo_go/trace"
)

type DiscardSink struct {
	*components.BasePlugin
	InStream components.Plugin
	counter  uint64
}

func NewDiscardSink(ctx context.Context, incoming components.Plugin, id int) (*DiscardSink, error) {
	ds := &DiscardSink{
		components.NewBasePlugin(id, nil),
		incoming,
		0,
	}
	return ds, nil
}

func (ds *DiscardSink) Run(ctx context.Context) error {
	for v := range ds.InStream.GetOutDataStream().Data {
		ds.counter += 1
		ds.doNothing(v)
		if ds.counter%1000 == 0 {
			log.Println("[DS-", ds.ID, "]Processed ", ds.counter, " incoming traces")
		}
	}

	return nil
}

func (ds *DiscardSink) doNothing(t *trace.ColumboTrace) {}

func (ds *DiscardSink) IncomingPlugins() []components.Plugin {
	return []components.Plugin{ds.InStream}
}
