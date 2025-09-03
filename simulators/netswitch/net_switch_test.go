package netswitch

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

		p, err := NewNetSwitchParser(context.Background(), 1, "netswitch_parser")
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

		assert.NoError(t, err)
	} else {
		fmt.Println("No log file provided! Skipping!")
	}
}
