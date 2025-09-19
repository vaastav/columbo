package main

import (
	"context"
	"flag"
	"log"
	"sync"

	"github.com/vaastav/columbo_go/parser"
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

	ctx := context.Background()

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
	// Set up the operator graph for all instances
	for _, instance := range sim_instances {
		// Create and launch a discard sink for each component
		for _, c := range instance.Components {
			wg.Add(1)
			ds, err := discard.NewDiscardSink(ctx, c.GetOutDataStream(), idx)
			if err != nil {
				log.Fatal(err)
			}
			go func(ds *discard.DiscardSink) {
				defer wg.Done()
				ds.Run(ctx)
			}(ds)
			idx += 1
		}
	}

	// Start the processing
	for sim_name, reader := range readers {
		wg.Add(1)
		go func(name string, reader *parser.Reader) {
			defer wg.Done()
			instance := sim_instances[sim_name]
			if instance != nil {
				err := reader.ProcessLog(ctx, instance.Process)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				log.Fatalf("No instance found for simulator %s\n", sim_name)
			}
		}(sim_name, reader)
	}

	wg.Wait()
	log.Println("Finished processing all traces")
}
