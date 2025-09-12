package components

import (
	"context"

	"github.com/vaastav/columbo_go/trace"
)

type DataStream struct {
	Data chan *trace.ColumboTrace
	Ctx  context.Context
}

func NewDataStream(ctx context.Context, data_channel chan *trace.ColumboTrace) (*DataStream, error) {
	return &DataStream{Data: data_channel, Ctx: ctx}, nil
}

func (ds *DataStream) Push(t *trace.ColumboTrace) {
	ds.Data <- t
}

func (ds *DataStream) Close() {
	// Close our channel
	close(ds.Data)
}
