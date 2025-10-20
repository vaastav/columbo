package materialize

import (
	"context"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vaastav/columbo_go/plugins/dummy"
	"github.com/vaastav/columbo_go/trace"
)

func TestSaveAndLoad(t *testing.T) {
	filename := "test_trace.gob"
	BUFFER_SIZE := 65536
	op := func(t *trace.ColumboTrace) bool {
		if t.Type == trace.TRACE {
			return true
		}
		return false
	}
	ctx := context.Background()
	ins, err := dummy.NewDummyGen(ctx, 0, 1000)
	require.NoError(t, err)
	s, err := NewSaver(ctx, ins, BUFFER_SIZE, 1, filename, op)
	require.NoError(t, err)
	l, err := NewLoader(ctx, BUFFER_SIZE, 2, filename)
	require.NoError(t, err)

	err = ins.Run(ctx)
	require.NoError(t, err)
	log.Println("Finished pushing traces from dummy generator")
	ins.GetOutDataStream().Close()
	err = s.Run(ctx)
	require.NoError(t, err)
	err = l.Run(ctx)
	require.NoError(t, err)
}
