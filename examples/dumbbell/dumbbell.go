package main

import (
	"context"
	"log"
	"sync"

	"github.com/vaastav/columbo_go/components"
	"github.com/vaastav/columbo_go/plugins/discard"
	"github.com/vaastav/columbo_go/plugins/export"
	"github.com/vaastav/columbo_go/plugins/filter"
	"github.com/vaastav/columbo_go/plugins/network"
	"github.com/vaastav/columbo_go/plugins/nic"
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

	sim_instances := simulation.Instances
	id := 0

	var wg sync.WaitGroup
	var sinks []components.Plugin
	var end_plugins []components.Plugin
	pairs, err := symphony.HostNicPairs(simulation)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Number of host-nic pairs", len(pairs))
	hosts, err := symphony.Hosts(simulation)
	if err != nil {
		log.Fatal(err)
	}
	for _, h := range hosts {
		ds, err := discard.NewDiscardSink(ctx, h, id)
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
		dma, err := nic.NewNicDmaTraceGen(ctx, n, BUFFER_SIZE, id)
		if err != nil {
			log.Fatal(err)
		}
		id++
		/*
			nictx, err := nic.NewNicTx(ctx, dma, BUFFER_SIZE, id)
			if err != nil {
				log.Fatal(err)
			}
			id++
		*/
		filter, err := filter.NewFilter(ctx, dma, BUFFER_SIZE, id, func(t *trace.ColumboTrace) bool {
			return t.Type == trace.TRACE || t.Type == trace.SPAN
		})
		id++
		end_plugins = append(end_plugins, filter)
	}
	switches, err := symphony.Switches(simulation)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Number of switches", len(switches))
	for _, sw := range switches {
		ntgen, err := network.NewNetworkTraceGen(ctx, sw, BUFFER_SIZE, id)
		if err != nil {
			log.Fatal(err)
		}
		id++
		//end_plugins = append(end_plugins, ntgen)
		ds, err := network.NewNetworkTraceGen(ctx, ntgen, BUFFER_SIZE, id)
		if err != nil {
			log.Fatal(err)
		}
		id++
		sinks = append(sinks, ds)
	}
	exp, err := export.NewExportSink(ctx, id, end_plugins)
	if err != nil {
		log.Fatal(err)
	}
	id++
	sinks = append(sinks, exp)
	switchpairs, err := symphony.SwitchPairs(simulation)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Number of switch-switch pairs", len(switchpairs))
	switchnicpairs, err := symphony.NicSwitchPairs(simulation)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Number of nic-switch pairs", len(switchnicpairs))

	symphony.LauncOpGraph(ctx, &wg, sinks)

	err = symphony.StartProcessing(readers, sim_instances)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Finished parsing all the files")
	wg.Wait()
	log.Println("Finished processing all files")
}
