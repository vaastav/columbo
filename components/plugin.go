package components

import (
	"context"
	"fmt"
	"log"
	"sync"
)

type Plugin interface {
	GetOutDataStream() *DataStream
	Shutdown(ctx context.Context) error
	Run(ctx context.Context) error
	IncomingPlugins() []Plugin
	IsLaunched() bool
	SetRunning(status bool)
}

type BasePlugin struct {
	ID        int
	OutStream *DataStream
	Running   bool
}

func (bp *BasePlugin) IsLaunched() bool {
	return bp.Running
}

func (bp *BasePlugin) GetOutDataStream() *DataStream {
	return bp.OutStream
}

func (bp *BasePlugin) Shutdown(context.Context) error { return nil }

func (bp *BasePlugin) Run(ctx context.Context) error {
	return fmt.Errorf("Run function is not implemented for the Base plugin %d", bp.ID)
}

func (bp *BasePlugin) IncomingPlugins() []Plugin {
	return []Plugin{}
}

func (bp *BasePlugin) SetRunning(status bool) {
	bp.Running = status
}

func NewBasePlugin(ID int, outs *DataStream) *BasePlugin {
	return &BasePlugin{ID: ID, OutStream: outs, Running: false}
}

func LaunchPlugin(ctx context.Context, wg *sync.WaitGroup, bp Plugin) {
	if bp.IsLaunched() {
		// We have already launched this plugin and its incoming plugins!
		return
	}
	// Launch all parent plugins
	for _, incoming := range bp.IncomingPlugins() {
		LaunchPlugin(ctx, wg, incoming)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := bp.Run(ctx)
		if err != nil {
			log.Println(err)
		}
	}()
	bp.SetRunning(true)
}
