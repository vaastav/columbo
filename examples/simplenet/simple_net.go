package main

import (
	"context"
	"log"
	"sync"

	"github.com/vaastav/columbo_go/components"
	"github.com/vaastav/columbo_go/plugins/discard"
	"github.com/vaastav/columbo_go/symphony"
)

const (
	BUFFER_SIZE = 65536
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create readers and set up the simulator instances
	readers, sim_instances, err := symphony.InitializeFromFile(ctx, BUFFER_SIZE)

	idx := 0

	var wg sync.WaitGroup
	var sinks []components.Plugin
	// Set up the operator graph for all instances
	for _, instance := range sim_instances {
		// Create and launch a discard sink for each component
		for _, c := range instance.Components {
			ds, err := discard.NewDiscardSink(ctx, c, idx)
			if err != nil {
				log.Fatal(err)
			}
			idx += 1
			sinks = append(sinks, ds)
		}
	}

	// Launch the OpGraph
	symphony.LauncOpGraph(ctx, &wg, sinks)

	// Launch the readers
	err = symphony.StartProcessing(readers, sim_instances)
	if err != nil {
		log.Fatal(err)
	}
	// Wait for the operator graph to finish
	wg.Wait()
	log.Println("Finished processing all traces")
}
