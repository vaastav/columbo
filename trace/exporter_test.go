package trace

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
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
	// Create 100 empty ColumboTraces
	for i := 0; i < 100; i++ {
		tr := &ColumboTrace{Graph: make(map[string][]string), Attributes: make(map[string]string)}
		tr.Export()
	}
	err = root_tracer.Shutdown(context.Background())
	require.NoError(t, err)
}
