package symphony

import (
	"context"
	"fmt"

	"github.com/vaastav/columbo_go/parser"
	"github.com/vaastav/columbo_go/topology"
)

func CreateReaders(ctx context.Context, lm *topology.LogMapping) (map[string]*parser.Reader, error) {
	readers := make(map[string]*parser.Reader)
	for sim, info := range lm.SimulatorLogs {
		var file_type parser.Source
		if info.FileType == "file" {
			file_type = parser.LogFile
		} else if info.FileType == "namedpipe" {
			file_type = parser.NamedPipe
		} else {
			return nil, fmt.Errorf("Invalid event source %s for the simulator %s\n", info.FileType, sim)
		}
		reader, err := parser.NewReader(ctx, info.Filename, file_type)
		if err != nil {
			return nil, err
		}
		readers[sim] = reader
	}
	return readers, nil
}
