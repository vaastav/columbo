package gem5

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

		p, err := NewGem5Parser(context.Background(), 1, "gem5_parser")
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

func TestInstrTrace(t *testing.T) {
	line := "857006011458: system.cpu: A0 T0 : 0xffffffff81600132 @syscall_return_via_sysret+67. 2 : POP_R : mov rsp, rsp, t1 : IntAlu : D=0x00007fff49f994c8 flags=(IsInteger|IsMicroop|IsLastMicroop)"
	p, err := NewGem5Parser(context.Background(), 1, "gem5_parser")
	require.NoError(t, err)
	event, err := p.ParseEvent(line)
	require.NotNil(t, event)
}
