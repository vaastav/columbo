package packet

import (
	"context"
	"sync"

	"github.com/vaastav/columbo_go/components"
	"github.com/vaastav/columbo_go/trace"
)

type PacketFull struct {
	*components.BasePlugin
	TxStream  components.Plugin
	NetStream components.Plugin
	RxStream  components.Plugin
	sendc     chan *trace.ColumboTrace
	netc      chan *trace.ColumboTrace
	recvc     chan *trace.ColumboTrace
}

func (p *PacketFull) mergeTraces(tx_t, net_t, rx_t *trace.ColumboTrace) *trace.ColumboTrace {
	t := &trace.ColumboTrace{}
	// TODO: Implement
	return t
}

func (p *PacketFull) Merge() {
	for {
		// Synchronize step!
		tx_t := <-p.sendc
		net_t := <-p.netc
		rx_t := <-p.recvc

		new_t := p.mergeTraces(tx_t, net_t, rx_t)
		p.OutStream.Push(new_t)
	}
}

func (p *PacketFull) processSend(ctx context.Context) {
	ins := p.TxStream.GetOutDataStream()
	for {
		select {
		case t := <-ins.Data:
			// Assumption: We are only getting send traces here
			// We can always check that tho

			// This is a blocking send on the channel. This will block until a receive is complete
			p.sendc <- t
		case <-ctx.Done():
			return
		}
	}
}

func (p *PacketFull) processNet(ctx context.Context) {
	ins := p.NetStream.GetOutDataStream()
	for {
		select {
		case t := <-ins.Data:
			// Assumption: We are only getting net traces here
			// We probably should check that tho

			// This is a blocking send on the channel. This will block until a receive is complete
			p.netc <- t
		case <-ctx.Done():
			return
		}
	}
}

func (p *PacketFull) processRecv(ctx context.Context) {
	ins := p.RxStream.GetOutDataStream()
	for {
		select {
		case t := <-ins.Data:
			// Assumption: We are only getting net traces here
			// We probably should check that tho

			// This is a blocking send on the channel. This will block until a receive is complete
			p.recvc <- t
		case <-ctx.Done():
			return
		}
	}
}

func (p *PacketFull) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	// Launch merging goroutine
	go func() {
		p.Merge()
	}()
	// Launch processing goroutines
	wg.Add(3)
	new_ctx, cancel := context.WithCancel(ctx)
	go func() {
		defer wg.Done()
		p.processSend(new_ctx)
	}()
	go func() {
		defer wg.Done()
		p.processRecv(new_ctx)
	}()
	go func() {
		defer wg.Done()
		p.processNet(new_ctx)
	}()
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-ctx.Done():
		cancel()
	}
	// Wait for everything to be wrapped up
	<-done
	return nil
}

func (p *PacketFull) IncomingPlugins() []components.Plugin {
	return []components.Plugin{p.TxStream, p.RxStream, p.TxStream}
}
