package trace

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vaastav/columbo_go/events"
)

var jaegerurl_ptr *string = flag.String("url", "http://localhost:14268/api/traces", "URL for jaeger collector")

func TestMain(m *testing.M) {
	flag.Parse()
	code := m.Run()
	os.Exit(code)
}

func TestExport(t *testing.T) {
	// Initialize the exporter
	fmt.Println("URL for collector is", *jaegerurl_ptr)
	err := InitJaegerExporter(*jaegerurl_ptr)
	require.NoError(t, err)
	// Create a few empty ColumboTraces
	/*
		for i := 0; i < 10; i++ {
			tr := &ColumboTrace{Graph: make(map[string][]string), Attributes: make(map[string]string)}
			tr.Export()
		}
	*/

	host1_tp, err := NewColumboTracer("host1")
	require.NoError(t, err)
	nic1_tp, err := NewColumboTracer("nic1")
	require.NoError(t, err)
	switch1_tp, err := NewColumboTracer("switch1")
	require.NoError(t, err)
	host_span := ColumboSpan{
		BaseTracer: host1_tp,
		Name:       "Doorbell write",
		ID:         "host1_mmio1",
		Events:     []events.Event{*events.NewEvent("host1_1", events.KHostMmioWT, 1000, 1, "gem5", ""), *events.NewEvent("host1_2", events.KHostMmioCWT, 1200, 1, "gem5", "")},
	}
	nic1_span := ColumboSpan{
		BaseTracer: nic1_tp,
		Name:       "Nic DMA read",
		ID:         "nic1_dma1",
		Events:     []events.Event{*events.NewEvent("nic1_1", events.KNicDmaT, 1100, 2, "i40", ""), *events.NewEvent("nic1_2", events.KNicDmaCRT, 1110, 2, "i40", "")},
	}
	sw1_span := ColumboSpan{
		BaseTracer: switch1_tp,
		Name:       "Packet tx",
		ID:         "switch_pkt1",
		Events:     []events.Event{*events.NewEvent("switch1_1", events.KNetworKEnqueueT, 1200, 3, "ns3", ""), *events.NewEvent("switch1_2", events.KNetworKDequeueT, 1220, 3, "ns3", "")},
	}

	// Setup different trace configs
	spans := []ColumboSpan{host_span, nic1_span, sw1_span}
	var traces []ColumboTrace
	for i := 0; i < 3; i++ {
		trace1 := ColumboTrace{
			Spans:      spans,
			Graph:      make(map[string][]string),
			Attributes: make(map[string]string),
		}
		traces = append(traces, trace1)
	}

	// 1 parent 2 children
	traces[0].Graph[sw1_span.ID] = []string{}
	traces[0].Graph[nic1_span.ID] = []string{}
	traces[0].Graph[host_span.ID] = []string{sw1_span.ID, nic1_span.ID}

	// 1 parent 1 child 1 grandchild
	traces[1].Graph[host_span.ID] = []string{nic1_span.ID}
	traces[1].Graph[nic1_span.ID] = []string{sw1_span.ID}
	traces[1].Graph[sw1_span.ID] = []string{}

	// 2 parents 1 child
	traces[2].Graph[host_span.ID] = []string{sw1_span.ID}
	traces[2].Graph[nic1_span.ID] = []string{sw1_span.ID}
	traces[2].Graph[sw1_span.ID] = []string{}

	for _, tr := range traces {
		tr.Export()
	}

	// Shutdown the trace provider so that it pushes all the spans
	err = host1_tp.Shutdown(context.Background())
	require.NoError(t, err)
	err = nic1_tp.Shutdown(context.Background())
	require.NoError(t, err)
	err = switch1_tp.Shutdown(context.Background())
	require.NoError(t, err)
	err = root_tracer.Shutdown(context.Background())
	require.NoError(t, err)
}
