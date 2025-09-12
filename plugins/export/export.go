package export

import (
	"context"
	"log"
	"sync"

	"github.com/vaastav/columbo_go/components"
)

type ExportSink struct {
	InStreams []*components.DataStream
}

func NewExportSink(ctx context.Context, instreams []*components.DataStream) (*ExportSink, error) {
	es := &ExportSink{
		InStreams: instreams,
	}
	return es, nil
}

func (es *ExportSink) Run(ctx context.Context) error {

	var wg sync.WaitGroup
	for _, stream := range es.InStreams {
		wg.Add(1)
		go func(ins *components.DataStream) {
			defer wg.Done()
			for v := range ins.Data {
				v.Export()
			}
		}(stream)
	}

	done := make(chan bool)

	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		log.Println("Finished exporting all traces")
		return nil
	case <-ctx.Done():
		log.Println("Context is done")
		return nil
	}
}
