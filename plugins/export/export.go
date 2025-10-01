package export

import (
	"context"
	"log"
	"sync"

	"github.com/vaastav/columbo_go/components"
)

type ExportSink struct {
	*components.BasePlugin
	InStreams []components.Plugin
}

func NewExportSink(ctx context.Context, ID int, instreams []components.Plugin) (*ExportSink, error) {
	es := &ExportSink{
		components.NewBasePlugin(ID, nil),
		instreams,
	}
	return es, nil
}

func (es *ExportSink) Run(ctx context.Context) error {

	var wg sync.WaitGroup
	for _, stream := range es.InStreams {
		wg.Add(1)
		go func(ins components.Plugin) {
			defer wg.Done()
			for v := range ins.GetOutDataStream().Data {
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

func (es *ExportSink) IncomingPlugins() []components.Plugin {
	return es.InStreams
}
