package symphony

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/vaastav/columbo_go/components"
	"github.com/vaastav/columbo_go/parser"
)

func StartProcessing(readers map[string]*parser.Reader, sim_instances map[string]*SimInstance) error {
	ctx := context.Background()
	var wg sync.WaitGroup
	// Start the processing
	for sim_name, reader := range readers {
		wg.Add(1)
		instance := sim_instances[sim_name]
		if instance == nil {
			return fmt.Errorf("No simulator instance found for the reader", sim_name)
		}
		go func(name string, reader *parser.Reader, instance *SimInstance) {
			defer wg.Done()
			err := reader.ProcessLog(ctx, instance.Process)
			if err != nil {
				log.Fatal(err)
			}
		}(sim_name, reader, instance)
	}

	// Wait for all parsers to finish
	wg.Wait()

	return nil
}

func LauncOpGraph(ctx context.Context, wg *sync.WaitGroup, sinks []components.Plugin) {
	for _, sink := range sinks {
		sink.Launch(ctx, wg)
	}
}
