package symphony

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"

	"github.com/vaastav/columbo_go/components"
	"github.com/vaastav/columbo_go/parser"
	"github.com/vaastav/columbo_go/topology"
)

var topology_file = flag.String("topology", "", "Path to topology file")
var logmapping = flag.String("log", "", "Path to log mapping file")

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
		components.LaunchPlugin(ctx, wg, sink)
	}
}

func Initialize(ctx context.Context, lm *topology.LogMapping, topo *topology.Topology, buffer_size int) (map[string]*parser.Reader, map[string]*SimInstance, error) {
	readers, err := CreateReaders(ctx, lm)
	if err != nil {
		return nil, nil, err
	}
	sim_instances, err := CreateSimInstanceFromTopology(ctx, topo, buffer_size)
	if err != nil {
		return nil, nil, err
	}
	return readers, sim_instances, nil
}

func InitializeFromFile(ctx context.Context, buffer_size int) (map[string]*parser.Reader, map[string]*SimInstance, error) {

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
		return nil, nil, err
	}

	lm, err := topology.ParseLogMapping(*logmapping)
	if err != nil {
		return nil, nil, err
	}
	return Initialize(ctx, lm, topo, buffer_size)
}
