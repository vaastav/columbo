package nicbm

import (
	"context"
	"errors"
	"regexp"
	"strconv"

	"github.com/vaastav/columbo_go/events"
)

type NicBMParser struct {
	Identifier int64
	Name       string
	id_cntr    uint64
}

func NewNicBMParser(ctx context.Context, identifier int64, name string) (*NicBMParser, error) {
	return &NicBMParser{Identifier: identifier, Name: name, id_cntr: 0}, nil
}

var (
	reMacAddr   = regexp.MustCompile(`info: mac_addr=([0-9a-fA-F]+)`)
	reSyncPci   = regexp.MustCompile(`sync_pci=(\d+) sync_eth=(\d+)`)
	reReadWrite = regexp.MustCompile(`main_time = (\d+): nicbm: (read|write)\(off=(0x[0-9a-fA-F]+), len=(\d+), val=(0x[0-9a-fA-F]+)(?:, posted=(\d+))?\)`)
	reDMAIssue  = regexp.MustCompile(`main_time = (\d+): nicbm: issuing dma op (\S+) addr (0x[0-9a-fA-F]+) len (\d+) pending (\d+)`)
	reDMAExec   = regexp.MustCompile(`main_time = (\d+): nicbm: executing dma op (\S+) addr (0x[0-9a-fA-F]+) len (\d+) pending (\d+)`)
	reDMAComp   = regexp.MustCompile(`main_time = (\d+): nicbm: completed dma (read|write) op (\S+) addr (0x[0-9a-fA-F]+) len (\d+)`)
	reDMAWData  = regexp.MustCompile(`main_time = (\d+): nicbm: dma write data`)
	reEthTx     = regexp.MustCompile(`main_time = (\d+): nicbm: eth tx: len (\d+)`)
	reEthRx     = regexp.MustCompile(`main_time = (\d+): nicbm: eth rx: port (\d+) len (\d+)`)
	reMSIIssue  = regexp.MustCompile(`main_time = (\d+): nicbm: issue (MSI(?:-X)?) interrupt vec (\d+)`)
)

func (p *NicBMParser) parseReadWriteEvent(m []string) (*events.Event, error) {
	id := strconv.FormatUint(p.id_cntr, 10)
	ts, err := strconv.ParseUint(m[1], 10, 64)
	if err != nil {
		return nil, err
	}
	var e_type events.EventType
	if m[2] == "read" {
		e_type = events.KNicMmioRT
	} else if m[2] == "write" {
		e_type = events.KNicMmioWT
	} else {
		return nil, errors.New("Invalid event operation for NIC: " + m[2])
	}
	off := m[3]
	len := m[4]
	val := m[5]
	e := events.NewEvent(id, e_type, ts, p.Identifier, p.Name, "")
	// Set all attributes
	e.AddAttribute("offset", off)
	e.AddAttribute("length", len)
	e.AddAttribute("val", val)
	if e_type == events.KNicMmioWT {
		posted := m[6]
		e.AddAttribute("posted", posted)
	}
	return e, nil
}

func (p *NicBMParser) parseDMAIssueEvent(m []string) (*events.Event, error) {
	id := strconv.FormatUint(p.id_cntr, 10)
	ts, err := strconv.ParseUint(m[1], 10, 64)
	if err != nil {
		return nil, err
	}
	e_type := events.KNicDmaIT
	op := m[2]
	addr := m[3]
	len := m[4]
	pending := m[5]
	e := events.NewEvent(id, e_type, ts, p.Identifier, p.Name, "")
	e.AddAttribute("op", op)
	e.AddAttribute("addr", addr)
	e.AddAttribute("len", len)
	e.AddAttribute("pending", pending)
	return e, nil
}

func (p *NicBMParser) parseDMAExecEvent(m []string) (*events.Event, error) {
	id := strconv.FormatUint(p.id_cntr, 10)
	ts, err := strconv.ParseUint(m[1], 10, 64)
	if err != nil {
		return nil, err
	}
	e_type := events.KNicDmaExT
	op := m[2]
	addr := m[3]
	len := m[4]
	pending := m[5]
	e := events.NewEvent(id, e_type, ts, p.Identifier, p.Name, "")
	e.AddAttribute("op", op)
	e.AddAttribute("addr", addr)
	e.AddAttribute("len", len)
	e.AddAttribute("pending", pending)
	return e, nil
}

func (p *NicBMParser) parseDMACompEvent(m []string) (*events.Event, error) {
	id := strconv.FormatUint(p.id_cntr, 10)
	ts, err := strconv.ParseUint(m[1], 10, 64)
	if err != nil {
		return nil, err
	}
	var e_type events.EventType
	if m[2] == "read" {
		e_type = events.KNicDmaCRT
	} else if m[2] == "write" {
		e_type = events.KNicDmaCWT
	} else {
		return nil, errors.New("Unsupported event type in dma completion " + m[2])
	}
	op := m[3]
	addr := m[4]
	len := m[5]
	e := events.NewEvent(id, e_type, ts, p.Identifier, p.Name, "")
	e.AddAttribute("op", op)
	e.AddAttribute("addr", addr)
	e.AddAttribute("len", len)
	return e, nil
}

func (p *NicBMParser) parseMacAddrEvent(m []string) (*events.Event, error) {
	id := strconv.FormatUint(p.id_cntr, 10)
	e := events.NewEvent(id, events.KEventT, 0, p.Identifier, p.Name, "")
	e.AddAttribute("mac_addr", m[1])
	return e, nil
}

func (p *NicBMParser) parseSyncPciEvent(m []string) (*events.Event, error) {
	id := strconv.FormatUint(p.id_cntr, 10)
	e := events.NewEvent(id, events.KEventT, 0, p.Identifier, p.Name, "")
	e.AddAttribute("sync_pci", m[1])
	e.AddAttribute("sync_eth", m[2])
	return e, nil
}

func (p *NicBMParser) parseDMAWriteData(m []string) (*events.Event, error) {
	id := strconv.FormatUint(p.id_cntr, 10)
	ts, err := strconv.ParseUint(m[1], 10, 64)
	if err != nil {
		return nil, err
	}
	e := events.NewEvent(id, events.KNicDmaWDataT, ts, p.Identifier, p.Name, "")
	// TODO: Correctly parse the write data once the source is fixed!
	return e, nil
}

func (p *NicBMParser) parseEthTxEvent(m []string) (*events.Event, error) {
	id := strconv.FormatUint(p.id_cntr, 10)
	ts, err := strconv.ParseUint(m[1], 10, 64)
	if err != nil {
		return nil, err
	}
	e := events.NewEvent(id, events.KNicTxT, ts, p.Identifier, p.Name, "")
	e.AddAttribute("len", m[2])
	return e, nil
}

func (p *NicBMParser) parseEthRxEvent(m []string) (*events.Event, error) {
	id := strconv.FormatUint(p.id_cntr, 10)
	ts, err := strconv.ParseUint(m[1], 10, 64)
	if err != nil {
		return nil, err
	}
	e := events.NewEvent(id, events.KNicRxT, ts, p.Identifier, p.Name, "")
	e.AddAttribute("port", m[2])
	e.AddAttribute("len", m[3])
	return e, nil
}

func (p *NicBMParser) parseMSIIssueEvent(m []string) (*events.Event, error) {
	id := strconv.FormatUint(p.id_cntr, 10)
	ts, err := strconv.ParseUint(m[1], 10, 64)
	if err != nil {
		return nil, err
	}
	e := events.NewEvent(id, events.KNicMsixT, ts, p.Identifier, p.Name, "")
	if m[2] == "MSI-X" {
		e.AddAttribute("isX", "true")
	} else {
		e.AddAttribute("isX", "false")
	}
	e.AddAttribute("vec", m[3])
	return e, nil
}

func (p *NicBMParser) ParseEvent(line string) (*events.Event, error) {
	if line == "" {
		return nil, nil
	}
	// Update the event ID counter
	p.id_cntr += 1
	if m := reReadWrite.FindStringSubmatch(line); m != nil {
		return p.parseReadWriteEvent(m)
	} else if m := reDMAIssue.FindStringSubmatch(line); m != nil {
		return p.parseDMAIssueEvent(m)
	} else if m := reDMAExec.FindStringSubmatch(line); m != nil {
		return p.parseDMAExecEvent(m)
	} else if m := reDMAComp.FindStringSubmatch(line); m != nil {
		return p.parseDMACompEvent(m)
	} else if m := reMacAddr.FindStringSubmatch(line); m != nil {
		return p.parseMacAddrEvent(m)
	} else if m := reSyncPci.FindStringSubmatch(line); m != nil {
		return p.parseSyncPciEvent(m)
	} else if m := reDMAWData.FindStringSubmatch(line); m != nil {
		return p.parseDMAWriteData(m)
	} else if m := reEthTx.FindStringSubmatch(line); m != nil {
		return p.parseEthTxEvent(m)
	} else if m := reEthRx.FindStringSubmatch(line); m != nil {
		return p.parseEthRxEvent(m)
	} else if m := reMSIIssue.FindStringSubmatch(line); m != nil {
		return p.parseMSIIssueEvent(m)
	}

	return nil, errors.New("Failed to parse log line `" + line + "`")
}
