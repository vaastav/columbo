package gem5

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/vaastav/columbo_go/events"
	"github.com/vaastav/columbo_go/parser"
)

type Gem5Parser struct {
	parser.BaseLogParser
	Identifier int64
	Name       string
	id_cntr    uint64
}

func NewGem5Parser(ctx context.Context, identifier int64, name string) (*Gem5Parser, error) {
	return &Gem5Parser{Identifier: identifier, Name: name, id_cntr: 0}, nil
}

// Top-level regex
var (
	reGlobal    = regexp.MustCompile(`(\d+): global: simbricks: (.+)`)
	reSystemCPU = regexp.MustCompile(`(\d+): system.cpu: (.+)`)
	reSystemPC  = regexp.MustCompile(`(\d+): system.pc.([\w\.]+): (.+)`)
)

var (
	// system.pc.pci_host regex
	rePCIHostRW = regexp.MustCompile(`([0-9a-f:]+.[0-9a-f]+): (read|write): offset=(0x[0-9a-f]+), size=(0x[0-9a-f]+)`)

	// system.pc.simbricks_0: regex
	rePCSim0Conf    = regexp.MustCompile(`(read|write)Config:\s+dev (\d+) func (\d+) reg (0|0x[0-9a-f]+) (\d+) bytes: data = (0|0x[0-9a-f]+)`)
	rePCSim0Send    = regexp.MustCompile(`simbricks-pci: sending (read|write) addr ([0-9a-f]+) size (\d+) id (\d+) bar (\d+) offs ([0-9a-f]+)(?: posted (\d+))?`)
	rePCSim0Recv    = regexp.MustCompile(`simbricks-pci: received (read|write) completion id (\d+)`)
	rePCSim0DMA     = regexp.MustCompile(`simbricks-pci: received DMA (read|write) id (\d+) addr ([0-9a-f]+) size ([0-9a-f]+)`)
	rePCSim0DMAComp = regexp.MustCompile(`simbricks-pci: completed DMA id (\d+)`)
	rePCSim0MSIX    = regexp.MustCompile(`simbricks-pci: received MSI-X intr vec (\d+)`)

	// system.pc.south_bridge.ide regex
	rePCSim0SBConf = regexp.MustCompile(`(read|write)Config:\s+dev (0x[0-9a-f]+) func (\d+) reg (0|0x[0-9a-f]+) (\d+) bytes: data = (0|0x[0-9a-f]+)`)
)

var (
	reCpuInstr   = regexp.MustCompile(`A(\d+)\s+T(\d+)\s+:\s+(0x[0-9a-fA-F]+)(?:\s+@(\w+)\+(\d+))?\s+:\s+(\w+)`)
	reCpuMicroop = regexp.MustCompile(`A(\d+)\s+T(\d+)\s+:\s+(0x[0-9a-fA-F]+)(?:\s+@(\w+)\+(\d+))?\.\s*(\d+)\s*:\s*(\w+)`)
)

func (p *Gem5Parser) parseGlobalEvent(m []string) (*events.Event, error) {
	id := strconv.FormatUint(p.id_cntr, 10)
	ts, err := strconv.ParseUint(m[1], 10, 64)
	if err != nil {
		return nil, err
	}
	var et events.EventType
	// Base event
	et = events.KEventT
	// Switch to a better suited event
	if m[2] == "processInEvent" {
		et = events.KSimProcInEventT
	}
	e := events.NewEvent(id, et, ts, p.Identifier, p.Name, m[2])
	return e, nil
}

func (p *Gem5Parser) parseCPUEvent(m []string) (*events.Event, error) {
	id := strconv.FormatUint(p.id_cntr, 10)
	ts, err := strconv.ParseUint(m[1], 10, 64)
	if err != nil {
		return nil, err
	}
	e := events.NewEvent(id, events.KHostInstrT, ts, p.Identifier, p.Name, m[2])
	if subm := reCpuInstr.FindStringSubmatch(m[2]); subm != nil {
		e.AddAttribute("cpu_id", subm[1])
		e.AddAttribute("tid", subm[2])
		e.AddAttribute("addr", subm[3])
		e.AddAttribute("type", "instr")
		e.AddAttribute("function", subm[4])
		e.AddAttribute("function_line", subm[5])
		e.AddAttribute("instr_name", subm[6])
		e.AddAttribute("exec_id", "A"+subm[1]+"T"+subm[2])
	} else if subm := reCpuMicroop.FindStringSubmatch(m[2]); subm != nil {
		e.AddAttribute("cpu_id", subm[1])
		e.AddAttribute("tid", subm[2])
		e.AddAttribute("addr", subm[3])
		e.AddAttribute("type", "microop")
		e.AddAttribute("function", subm[4])
		e.AddAttribute("function_line", subm[5])
		e.AddAttribute("instr_name", subm[6])
		is_last_op := strings.Contains(m[2], "IsLastMicroop")
		e.AddAttribute("is_last_microop", strconv.FormatBool(is_last_op))
		e.AddAttribute("exec_id", "A"+subm[1]+"T"+subm[2])
	} else {
		return nil, fmt.Errorf("Unsupported instruction: %s", m[2])
	}
	return e, nil
}

func (p *Gem5Parser) parsePCEvent(m []string) (*events.Event, error) {
	id := strconv.FormatUint(p.id_cntr, 10)
	ts, err := strconv.ParseUint(m[1], 10, 64)
	if err != nil {
		return nil, err
	}
	op := m[2]
	msg := m[3]
	if op == "south_bridge.cmos.rtc" {
		e := events.NewEvent(id, events.KEventT, ts, p.Identifier, p.Name, msg)
		return e, nil
	} else if op == "pci_host" {
		if subm := rePCIHostRW.FindStringSubmatch(msg); subm != nil {
			pci_dev := subm[1]
			op := subm[2]
			offset := subm[3]
			size := subm[4]
			e := events.NewEvent(id, events.KHostPciRWT, ts, p.Identifier, p.Name, "")
			e.AddAttribute("pci_device", pci_dev)
			e.AddAttribute("op", op)
			e.AddAttribute("offset", offset)
			e.AddAttribute("size", size)
			return e, nil
		}
		e := events.NewEvent(id, events.KEventT, ts, p.Identifier, p.Name, msg)
		return e, nil
	} else if op == "simbricks_0" {
		if subm := rePCSim0Conf.FindStringSubmatch(msg); subm != nil {
			e := events.NewEvent(id, events.KHostConfT, ts, p.Identifier, p.Name, "")
			e.AddAttribute("type", subm[1])
			e.AddAttribute("dev", subm[2])
			e.AddAttribute("func", subm[3])
			e.AddAttribute("reg", subm[4])
			e.AddAttribute("bytes", subm[5])
			e.AddAttribute("data", subm[6])
			e.AddAttribute("source", op)
			return e, nil
		} else if msg == "simbricks-pci: device constructed" {
			e := events.NewEvent(id, events.KEventT, ts, p.Identifier, p.Name, msg)
			return e, nil
		} else if subm := rePCSim0Send.FindStringSubmatch(msg); subm != nil {
			var et events.EventType
			var posted string
			if subm[1] == "read" {
				et = events.KHostMmioRT
			} else if subm[1] == "write" {
				et = events.KHostMmioWT
				posted = subm[7]
			} else {
				return nil, errors.New("MMIO only supports read|write operations. Found invalid op: " + subm[1])
			}
			e := events.NewEvent(id, et, ts, p.Identifier, p.Name, "")
			e.AddAttribute("addr", subm[2])
			e.AddAttribute("size", subm[3])
			e.AddAttribute("id", subm[4])
			e.AddAttribute("bar", subm[5])
			e.AddAttribute("offs", subm[6])
			e.AddAttribute("posted", posted)
			return e, nil
		} else if subm := rePCSim0Recv.FindStringSubmatch(msg); subm != nil {
			var et events.EventType
			if subm[1] == "read" {
				et = events.KHostMmioCRT
			} else if subm[1] == "write" {
				et = events.KHostMmioCWT
			} else {
				return nil, errors.New("MMIO only supports read|write operations. Found invalid op: " + subm[1])
			}
			e := events.NewEvent(id, et, ts, p.Identifier, p.Name, "")
			e.AddAttribute("ID", subm[2])
			return e, nil
		} else if subm := rePCSim0DMA.FindStringSubmatch(msg); subm != nil {
			var et events.EventType
			if subm[1] == "read" {
				et = events.KHostDmaRT
			} else if subm[1] == "write" {
				et = events.KHostDmaWT
			} else {
				return nil, errors.New("DMA only supports read|write operations. Found invalid op: " + subm[1])
			}
			e := events.NewEvent(id, et, ts, p.Identifier, p.Name, "")
			e.AddAttribute("id", subm[2])
			e.AddAttribute("addr", subm[3])
			e.AddAttribute("size", subm[4])
			return e, nil
		} else if subm := rePCSim0DMAComp.FindStringSubmatch(msg); subm != nil {
			e := events.NewEvent(id, events.KHostDmaCT, ts, p.Identifier, p.Name, "")
			e.AddAttribute("id", subm[1])
			return e, nil
		} else if subm := rePCSim0MSIX.FindStringSubmatch(msg); subm != nil {
			e := events.NewEvent(id, events.KHostMsiXT, ts, p.Identifier, p.Name, "")
			e.AddAttribute("vec", subm[1])
			return e, nil
		}
	} else if op == "south_bridge.ide" {
		if subm := rePCSim0SBConf.FindStringSubmatch(msg); subm != nil {
			e := events.NewEvent(id, events.KHostConfT, ts, p.Identifier, p.Name, "")
			e.AddAttribute("type", subm[1])
			e.AddAttribute("dev", subm[2])
			e.AddAttribute("func", subm[3])
			e.AddAttribute("reg", subm[4])
			e.AddAttribute("bytes", subm[5])
			e.AddAttribute("data", subm[6])
			e.AddAttribute("source", op)
			return e, nil
		}
	} else if op == "simbricks_0.pio" {
		e := events.NewEvent(id, events.KEventT, ts, p.Identifier, p.Name, msg)
		return e, nil
	}
	log.Println(id, ts, op, msg)
	log.Println(op)
	return nil, errors.New("Not implemented")
}

func (p *Gem5Parser) ParseEvent(line string) (*events.Event, error) {
	if line == "" {
		return nil, nil
	}
	p.id_cntr += 1
	if m := reGlobal.FindStringSubmatch(line); m != nil {
		return p.parseGlobalEvent(m)
	} else if m := reSystemPC.FindStringSubmatch(line); m != nil {
		return p.parsePCEvent(m)
	} else if m := reSystemCPU.FindStringSubmatch(line); m != nil {
		return p.parseCPUEvent(m)
	}

	return nil, errors.New("Failed to parse log line `" + line + "`")
}
