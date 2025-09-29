package main

import (
	"context"
	"log"
	"sync"

	"github.com/vaastav/columbo_go/components"
	"github.com/vaastav/columbo_go/plugins/discard"
	"github.com/vaastav/columbo_go/plugins/network"
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
	id := 0

	var wg sync.WaitGroup
	var sinks []components.Plugin
	pairs, err := symphony.HostNicPairs(simulation)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Number of pairs", len(pairs))
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
		ds, err := discard.NewDiscardSink(ctx, ntgen, id)
		if err != nil {
			log.Fatal(err)
		}
		id++
		sinks = append(sinks, ds)
	}
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
