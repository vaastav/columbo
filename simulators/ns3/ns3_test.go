package ns3

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
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
		require.NoError(t, err)
		defer file.Close()

		p, err := NewNS3Parser(context.Background(), 1, "ns3_parser")
		require.NoError(t, err)

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if scanner.Text() == "" {
				continue
			}
			event, err := p.ParseEvent(scanner.Text())
			require.NoError(t, err)
			if event == nil {
				t.Log(scanner.Text())
			}
			require.NotNil(t, event)
		}
		require.NoError(t, err)
	} else {
		fmt.Println("No log file provided! Skipping!")
	}
}
