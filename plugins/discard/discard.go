package discard

import (
	"context"
	"log"

	"github.com/vaastav/columbo_go/components"
	"github.com/vaastav/columbo_go/trace"
)

type DiscardSink struct {
	InStream *components.DataStream
	ID       int
	counter  uint64
}

func NewDiscardSink(ctx context.Context, instream *components.DataStream, id int) (*DiscardSink, error) {
	ds := &DiscardSink{
		InStream: instream,
		ID:       id,
	}
	return ds, nil
}

func (ds *DiscardSink) Run(ctx context.Context) error {
	log.Println("[DS-", ds.ID, "]Run function start")
	for v := range ds.InStream.Data {
		ds.counter += 1
		ds.doNothing(v)
		if ds.counter%1000 == 0 {
			log.Println("[DS-", ds.ID, "]Processed ", ds.counter, " incoming traces")
		}
	}

	log.Println("[DS-", ds.ID, "]Finished processing all events")

	return nil
}

func (ds *DiscardSink) doNothing(t *trace.ColumboTrace) {}
