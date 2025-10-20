package materialize

import (
	"context"
	"encoding/gob"
	"errors"
	"log"
	"os"

	"github.com/vaastav/columbo_go/components"
	"github.com/vaastav/columbo_go/trace"
)

type Save struct {
	*components.BasePlugin
	Filename string
	Op       func(t *trace.ColumboTrace) bool
	InStream components.Plugin
	Traces   []*trace.ColumboTrace
}

func NewSaver(ctx context.Context, ins components.Plugin, buffer_size int, ID int, filename string, Op func(t *trace.ColumboTrace) bool) (*Save, error) {
	outs, err := components.NewDataStream(ctx, make(chan *trace.ColumboTrace, buffer_size))
	if err != nil {
		return nil, err
	}

	s := &Save{
		components.NewBasePlugin(ID, outs),
		filename,
		Op,
		ins,
		[]*trace.ColumboTrace{},
	}
	return s, nil
}

func (s *Save) SaveData() error {
	file, err := os.Create(s.Filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(s.Traces); err != nil {
		return err
	}
	log.Println("Number of saved traces: ", len(s.Traces))
	return nil
}

func (s *Save) Run(ctx context.Context) error {
	ds := s.InStream.GetOutDataStream()
	if ds == nil {
		return errors.New("Outdatastream of incoming plugin is nil")
	}
	for {
		select {
		case t, ok := <-ds.Data:
			if !ok {
				s.OutStream.Close()
				return s.SaveData()
			}
			if s.Op(t) {
				s.Traces = append(s.Traces, t)
			} else {
				s.OutStream.Push(t)
			}
		case <-ctx.Done():
			log.Println("Context is done. Quitting!")
			s.OutStream.Close()
			return s.SaveData()
		}
	}
}

func (s *Save) IncomingPlugins() []components.Plugin {
	return []components.Plugin{s.InStream}
}
