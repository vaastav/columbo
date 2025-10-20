package materialize

import (
	"context"
	"encoding/gob"
	"log"
	"os"

	"github.com/vaastav/columbo_go/components"
	"github.com/vaastav/columbo_go/trace"
)

type Load struct {
	*components.BasePlugin
	Filename string
}

func NewLoader(ctx context.Context, buffer_size int, ID int, filename string) (*Load, error) {
	outs, err := components.NewDataStream(ctx, make(chan *trace.ColumboTrace, buffer_size))
	if err != nil {
		return nil, err
	}

	l := &Load{
		components.NewBasePlugin(ID, outs),
		filename,
	}
	return l, nil
}

func (l *Load) LoadData() ([]*trace.ColumboTrace, error) {
	var res []*trace.ColumboTrace
	file, err := os.Open(l.Filename)
	if err != nil {
		return res, nil
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&res); err != nil {
		return res, err
	}
	log.Println("Number of loaded traces: ", len(res))
	return res, nil
}

func (l *Load) Run(ctx context.Context) error {
	traces, err := l.LoadData()
	if err != nil {
		return err
	}
	for _, t := range traces {
		l.OutStream.Push(t)
	}
	return nil
}

func (l *Load) IncomingPlugins() []components.Plugin {
	return []components.Plugin{}
}
