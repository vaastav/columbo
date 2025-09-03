package netswitch

import (
	"context"
	"errors"
	"regexp"
	"strconv"

	"github.com/vaastav/columbo_go/events"
)

type NetSwitchParser struct {
	Identifier int64
	Name       string
	id_cntr    uint64
}

func NewNetSwitchParser(ctx context.Context, identifier int64, name string) (*NetSwitchParser, error) {
	return &NetSwitchParser{Identifier: identifier, Name: name, id_cntr: 0}, nil
}

var (
	rePktTx   = regexp.MustCompile(`info: main_time=(\d+) Packet Transmit: id=(\d+) e_port=(\d+) i_port=(\d+)`)
	rePktDrop = regexp.MustCompile(`info: main_time=(\d+) Packet Drop: id=(\d+) e_port=(\d+) i_port=(\d+)`)
	rePktRx   = regexp.MustCompile(`info: main_time=(\d+) Packet Receive: id=(\d+) port=(\d+) src=([0-9a-f:]+) dst=([0-9a-f:]+) len=(\d+)`)
)

func (p *NetSwitchParser) parseRxEvent(m []string) (*events.Event, error) {
	id := strconv.FormatUint(p.id_cntr, 10)
	ts, err := strconv.ParseUint(m[1], 10, 64)
	if err != nil {
		return nil, err
	}
	e := events.NewEvent(id, events.KNetworKEnqueueT, ts, p.Identifier, p.Name, "")
	e.AddAttribute("transient_id", m[2])
	e.AddAttribute("ingress_port", m[3])
	e.AddAttribute("src_mac", m[4])
	e.AddAttribute("dst_mac", m[5])
	e.AddAttribute("len", m[6])
	return e, nil
}

func (p *NetSwitchParser) parseTxEvent(m []string) (*events.Event, error) {
	id := strconv.FormatUint(p.id_cntr, 10)
	ts, err := strconv.ParseUint(m[1], 10, 64)
	if err != nil {
		return nil, err
	}
	e := events.NewEvent(id, events.KNetworKDequeueT, ts, p.Identifier, p.Name, "")
	e.AddAttribute("transient_id", m[2])
	e.AddAttribute("egress_port", m[3])
	e.AddAttribute("ingress_port", m[4])
	return e, nil
}

func (p *NetSwitchParser) parseDropEvent(m []string) (*events.Event, error) {
	id := strconv.FormatUint(p.id_cntr, 10)
	ts, err := strconv.ParseUint(m[1], 10, 64)
	if err != nil {
		return nil, err
	}
	e := events.NewEvent(id, events.KNetworKDropT, ts, p.Identifier, p.Name, "")
	e.AddAttribute("transient_id", m[2])
	e.AddAttribute("egress_port", m[3])
	e.AddAttribute("ingress_port", m[4])
	return e, nil
}

func (p *NetSwitchParser) ParseEvent(line string) (*events.Event, error) {
	if line == "" {
		return nil, nil
	}
	p.id_cntr += 1
	if m := rePktRx.FindStringSubmatch(line); m != nil {
		return p.parseRxEvent(m)
	} else if m := rePktTx.FindStringSubmatch(line); m != nil {
		return p.parseTxEvent(m)
	} else if m := rePktDrop.FindStringSubmatch(line); m != nil {
		return p.parseDropEvent(m)
	}
	return nil, errors.New("Failed to parse log line `" + line + "`")
}
