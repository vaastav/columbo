package ns3

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/vaastav/columbo_go/events"
	"github.com/vaastav/columbo_go/parser"
)

type NS3Parser struct {
	parser.BaseLogParser
	id_cntr uint64
}

// Overall log pattern
var reLogPattern = regexp.MustCompile(`^\+([0-9.]+)s\s+(-?\d+)\s+([A-Za-z0-9_:]+)\((.*?)\)(.*)$`)

// Probable Packet Storyline

// BridgeNetDevice::ReceiveFromDevice()
// BridgeNetDevice::ReceiveFromDevice(): [DEBUG] UID is pkt_id
var reReceive = regexp.MustCompile(`: \[DEBUG\] UID is (\d+)`)

// BridgeNetDevice::ForwardBroadcast( LearningBridgeForward (src => dst): src_device_type --> dst_device_type (UID pkt_id))
// BridgeNetDevice::ForwardUnicast( LearningBridgeForward (src => dst): src_device_type --> dst_device_type (UID pkt_id) )
var reCastPattern = regexp.MustCompile(`^\s*:\s+\[DEBUG\]\s+(\w+)\s+\(incomingPort=([^,]+),\s+packet=([^,]+),\s+protocol=(\d+),\s+src=([^,]+),\s+dst=([^)]+)\)$`)

// SimBricksNetDevice:SendFrom(netdevice_addr, packet_addr, source, dst, protocol_number)
//    protocol_number is 2054 if ARP packet
//    protocol number is 2048 if IP packet

func NewNS3Parser(ctx context.Context, identifier int64, name string) (*NS3Parser, error) {
	return &NS3Parser{parser.BaseLogParser{Identifier: identifier, Name: name}, 0}, nil
}

func (p *NS3Parser) parseSend(id string, msg string, ts uint64, devid string) (*events.Event, error) {
	parts := strings.Split(msg, ",")
	if len(parts) != 5 {
		return nil, errors.New("SimbricksNetDevice SendFrom missing fields")
	}
	e := events.NewEvent(id, events.KNetworKDequeueT, ts, p.Identifier, p.Name, "")
	e.Attributes["method"] = "SendFrom"
	e.Attributes["class"] = "SimbricksNetDevice"
	e.Attributes["devid"] = devid
	e.Attributes["device_addr"] = parts[0]
	e.Attributes["pkt_addr"] = parts[1]
	e.Attributes["src_mac"] = parts[2]
	e.Attributes["dst_mac"] = parts[3]
	e.Attributes["protocol"] = get_protocol(parts[4])
	return e, nil
}

func (p *NS3Parser) parseReceive(id string, msg string, ts uint64, devid string) (*events.Event, error) {
	if msg == "" {
		e := events.NewEvent(id, events.KEventT, ts, p.Identifier, p.Name, "")
		e.Attributes["devid"] = devid
		e.Attributes["method"] = "ReceiveFromDevice"
		e.Attributes["class"] = "BridgeNetDevice"
		return e, nil
	} else {
		if m := reReceive.FindStringSubmatch(msg); m != nil {
			pkt_id := m[1]
			e := events.NewEvent(id, events.KNetworKEnqueueT, ts, p.Identifier, p.Name, "")
			e.Attributes["devid"] = devid
			e.Attributes["method"] = "ReceiveFromDevice"
			e.Attributes["class"] = "BridgeNetDevice"
			e.Attributes["transient_id"] = pkt_id
			return e, nil
		} else {
			return nil, errors.New("ReceiveFromDevice did not have a packet UID!")
		}
	}
}

func get_protocol(pid string) string {
	if pid == "2054" {
		return "ARP"
	} else if pid == "2048" {
		return "IP"
	}
	return "UNKNOWN"
}

func (p *NS3Parser) parseCast(id string, msg string, ts uint64, devid string, cast_type string) (*events.Event, error) {
	if m := reCastPattern.FindStringSubmatch(msg); m != nil {
		e := events.NewEvent(id, events.KEventT, ts, p.Identifier, p.Name, "")
		e.Attributes["devid"] = devid
		e.Attributes["method"] = "Forward" + cast_type
		e.Attributes["class"] = "BridgeNetDevice"
		e.Attributes["incomingPort"] = m[1]
		e.Attributes["packet"] = m[2]
		e.Attributes["protocol"] = get_protocol(m[3])
		e.Attributes["src_mac"] = m[4]
		e.Attributes["dst_mac"] = m[5]
		return e, nil
	} else if strings.HasPrefix(msg, ": [LOGIC]") {
		e := events.NewEvent(id, events.KEventT, ts, p.Identifier, p.Name, "")
		e.Attributes["devid"] = devid
		e.Attributes["method"] = "Forward" + cast_type
		e.Attributes["class"] = "BridgeNetDevice"
		e.Attributes["logic"] = msg
		return e, nil
	}
	return nil, errors.New("Unable to parse a cast message: " + msg)
}

func (p *NS3Parser) ParseEvent(line string) (*events.Event, error) {
	if line == "" {
		return nil, nil
	}
	p.id_cntr += 1
	id := strconv.FormatUint(p.id_cntr, 10)
	if m := reLogPattern.FindStringSubmatch(line); m != nil {
		ts, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			return nil, err
		}
		ts_int := uint64(ts * 1e12)
		devid := m[2]
		classMethod := m[3]
		class := classMethod
		method := ""
		if idx := regexp.MustCompile(`:`).FindStringIndex(classMethod); idx != nil {
			parts := regexp.MustCompile(`:`).Split(classMethod, 2)
			class, method = parts[0], parts[1]
		}
		if class == "SimbricksNetDevice" {
			if method == "SendFrom" {
				return p.parseSend(id, m[4], ts_int, devid)
			}
		} else if class == "BridgeNetDevice" {
			if method == "ReceiveFromDevice" {
				return p.parseReceive(id, m[5], ts_int, devid)
			} else if method == "ForwardBroadcast" && m[5] != "" {
				return p.parseCast(id, m[5], ts_int, devid, "Broadcast")
			} else if method == "ForwardUnicast" && m[5] != "" {
				return p.parseCast(id, m[5], ts_int, devid, "Unicast")
			}
		}

		// Parse the rest as just generic events
		e := events.NewEvent(id, events.KEventT, ts_int, p.Identifier, p.Name, m[5])
		e.AddAttribute("class", class)
		e.AddAttribute("method", method)
		e.AddAttribute("devid", devid)
		return e, nil
	} else if strings.Contains(line, "DoDispose()") {
		e := events.NewEvent(id, events.KEventT, 0, p.Identifier, p.Name, line)
		return e, nil
	}

	return nil, errors.New("Failed to parse log line `" + line + "`")
}
