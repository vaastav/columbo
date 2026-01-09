package main

import (
	"context"
	"log"
	"sync"

	"github.com/vaastav/columbo_go/components"
	"github.com/vaastav/columbo_go/plugins/discard"
	"github.com/vaastav/columbo_go/plugins/filter"
	"github.com/vaastav/columbo_go/plugins/host"
	"github.com/vaastav/columbo_go/symphony"
	"github.com/vaastav/columbo_go/trace"
)

const (
	BUFFER_SIZE = 65536
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := trace.InitJaegerExporter("http://localhost:14268/api/traces")
	if err != nil {
		log.Fatal(err)
	}

	readers, simulation, err := symphony.InitializeFromFile(ctx, BUFFER_SIZE)
	if err != nil {
		log.Fatal(err)
	}

	sim_instances := simulation.Instances
	id := 0

	var wg sync.WaitGroup
	var sinks []components.Plugin
	hosts, err := symphony.Hosts(simulation)
	if err != nil {
		log.Fatal(err)
	}
	for _, h := range hosts {
		syscall, err := host.NewSyscall(ctx, h, BUFFER_SIZE, id)
		if err != nil {
			log.Fatal(err)
		}
		id++
		filter, err := filter.NewFilter(ctx, syscall, BUFFER_SIZE, id, func(t *trace.ColumboTrace) bool {
			return t.Type == trace.SPAN
		})
		ds, err := discard.NewDiscardSink(ctx, filter, id)
		id++
		if err != nil {
			log.Fatal(err)
		}
		sinks = append(sinks, ds)
	}
	nics, err := symphony.NICs(simulation)
	if err != nil {
		log.Fatal(err)
	}
	for _, n := range nics {
		ds, err := discard.NewDiscardSink(ctx, n, id)
		if err != nil {
			log.Fatal(err)
		}
		id++
		sinks = append(sinks, ds)
	}
	switches, err := symphony.Switches(simulation)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Number of switches", len(switches))
	for _, sw := range switches {
		//end_plugins = append(end_plugins, ntgen)
		ds, err := discard.NewDiscardSink(ctx, sw, id)
		if err != nil {
			log.Fatal(err)
		}
		id++
		sinks = append(sinks, ds)
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
