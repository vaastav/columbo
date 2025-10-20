package dummy

import (
	"context"

	"github.com/vaastav/columbo_go/components"
	"github.com/vaastav/columbo_go/events"
	"github.com/vaastav/columbo_go/trace"
)

type DummyGen struct {
	*components.BasePlugin
	NumTraces int64
}

func NewDummyGen(ctx context.Context, ID int, NumTraces int64) (*DummyGen, error) {
	outs, err := components.NewDataStream(ctx, make(chan *trace.ColumboTrace, NumTraces))
	if err != nil {
		return nil, err
	}
	dummy_gen := &DummyGen{
		components.NewBasePlugin(ID, outs),
		NumTraces,
	}
	return dummy_gen, nil
}

func (d *DummyGen) genTrace() *trace.ColumboTrace {
	t := &trace.ColumboTrace{}
	t.Type = trace.TRACE
	t.Attributes = make(map[string]string)
	t.Attributes["source"] = "DUMMY GENERATOR"
	span := trace.ColumboSpan{}
	span.Events = append(span.Events, *events.NewEvent("0", events.KEventT, 0, 1, "dummy", "dummy event"))
	span.ID = "1"
	t.Graph = make(map[string][]string)
	t.Graph["1"] = []string{}
	return t
}

func (d *DummyGen) Run(ctx context.Context) error {
	for i := int64(0); i < d.NumTraces; i++ {
		t := d.genTrace()
		d.OutStream.Push(t)
	}
	return nil
}

func (d *DummyGen) IncomingPlugins() []components.Plugin {
	return []components.Plugin{}
}
