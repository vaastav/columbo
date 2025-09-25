package components

import "context"

type Plugin interface {
	GetOutDataStream() *DataStream
	Shutdown()
	Run(ctx context.Context) error
}
