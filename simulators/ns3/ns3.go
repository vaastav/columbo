package ns3

import (
	"context"

	"github.com/vaastav/columbo_go/parser"
)

type NS3Parser struct {
	parser.BaseLogParser
	id_cntr uint64
}

func NewNS3Parser(ctx context.Context, identifier int64, name string) (*NS3Parser, error) {
	return &NS3Parser{parser.BaseLogParser{Identifier: identifier, Name: name}, 0}, nil
}
