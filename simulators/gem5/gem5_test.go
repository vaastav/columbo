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

func TestInstr(t *testing.T) {
	p, err := NewGem5Parser(context.Background(), 1, "gem5_parser")
	require.NoError(t, err)
	// Main instr
	line := "857006012457: system.cpu: A0 T0 : 0xffffffff81600136 @syscall_return_via_sysret+71 : sysret"
	event, err := p.ParseEvent(line)
	require.NotNil(t, event)
	require.Len(t, event.Attributes, 7)
	// Microop
	line = "857006011458: system.cpu: A0 T0 : 0xffffffff81600132 @syscall_return_via_sysret+67. 2 : POP_R : mov rsp, rsp, t1 : IntAlu : D=0x00007fff49f994c8 flags=(IsInteger|IsMicroop|IsLastMicroop)"
	event, err = p.ParseEvent(line)
	require.NotNil(t, event)
	require.Len(t, event.Attributes, 8)

	// Main instr with no function name
	line = "857006012790: system.cpu: A0 T0 : 0x7fc14f71f8b0 : cmp rax, 0xfffffffffffff000"
	event, err = p.ParseEvent(line)
	require.NotNil(t, event)
	require.Len(t, event.Attributes, 7)

	// Microops with no function name (along with first and last microop flags)
	line = "857006012790: system.cpu: A0 T0 : 0x7fc14f71f8b0. 0 : CMP_R_I : limm t1, 0xfffffffffffff000 : IntAlu : D=0xfffffffffffff000 flags=(IsInteger|IsMicroop|IsDelayedCommit|IsFirstMicroop)"
	event, err = p.ParseEvent(line)
	require.NotNil(t, event)
	require.Len(t, event.Attributes, 8)
}
