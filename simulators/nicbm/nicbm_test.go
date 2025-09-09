package nicbm

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vaastav/columbo_go/events"
)

var sampleLogFile_ptr *string = flag.String("logfile", "", "Path to log file")

func TestMain(m *testing.M) {
	flag.Parse()
	code := m.Run()
	os.Exit(code)
}

func TestFile(t *testing.T) {
	t.Log("Sample log file is : ", *sampleLogFile_ptr)
	if *sampleLogFile_ptr != "" {
		file, err := os.Open(*sampleLogFile_ptr)
		assert.NoError(t, err)
		defer file.Close()

		p, err := NewNicBMParser(context.Background(), 1, "nicbm_parser")
		assert.NoError(t, err)

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if scanner.Text() == "" {
				continue
			}
			event, err := p.ParseEvent(scanner.Text())
			assert.NoError(t, err)
			if event == nil {
				t.Log(scanner.Text())
			}
			assert.NotNil(t, event)
		}

		assert.NoError(t, scanner.Err())
	} else {
		fmt.Println("No log file provided! Skipping!")
	}
	// Auto-pass the test if no file is provided
}

func TestDMAWriteData(t *testing.T) {
	p, err := NewNicBMParser(context.Background(), 1, "nicbm_parser")
	require.NoError(t, err)

	event, err := p.ParseEvent("info: main_time = 728327065000: nicbm: dma write data: 03 20 01 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 06 00 00 00 01 00 0A 00")
	assert.Equal(t, event.Type, events.KNicDmaWDataT)
	assert.Equal(t, "03 20 01 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 06 00 00 00 01 00 0A 00", event.Attributes["data"])
}
