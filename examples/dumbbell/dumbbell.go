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

	readers, simulation, err := symphony.InitializeFromFile(ctx, BUFFER_SIZE)

	sim_instances := simulation.Instances
	idx := 0

	var wg sync.WaitGroup
	var sinks []components.Plugin
	for _, instance := range sim_instances {
		for _, c := range instance.Components {
			ds, err := discard.NewDiscardSink(ctx, c, idx)
			if err != nil {
				log.Fatal(err)
			}
			idx += 1
			sinks = append(sinks, ds)
		}
	}

	symphony.LauncOpGraph(ctx, &wg, sinks)

	err = symphony.StartProcessing(readers, sim_instances)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Finished parsing all the files")
	wg.Wait()
	log.Println("Finished processing all files")
}
