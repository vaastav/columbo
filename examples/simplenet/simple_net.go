package main

import (
	"context"
	"flag"
	"log"
	"sync"

	"github.com/vaastav/columbo_go/components"
	"github.com/vaastav/columbo_go/plugins/discard"
	"github.com/vaastav/columbo_go/symphony"
	"github.com/vaastav/columbo_go/topology"
)

const (
	BUFFER_SIZE = 65536
)

var topology_file = flag.String("topology", "", "Path to topology file")
var logmapping = flag.String("log", "", "Path to log mapping file")

func main() {
	flag.Parse()
	if *topology_file == "" {
		log.Println("topology file for the simulation not provided")
		log.Fatal("Usage: go run simple_net.go -topology=path/to/topology.json -log=path/to/log_mapping.json")
	}

	if *logmapping == "" {
		log.Println("Log mapping file for the simulation not provided")
		log.Fatal("Usage: go run simple_net.go -topology=path/to/topology.json -log=path/to/log_mapping.json")
	}

	topo, err := topology.ParseTopology(*topology_file)
	if err != nil {
		log.Fatal(err)
	}

	lm, err := topology.ParseLogMapping(*logmapping)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create readers and set up the simulator instances
	readers, err := symphony.CreateReaders(ctx, lm)
	if err != nil {
		log.Fatal(err)
	}

	sim_instances, err := symphony.CreateSimInstanceFromTopology(ctx, topo, BUFFER_SIZE)
	if err != nil {
		log.Fatal(err)
	}

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
