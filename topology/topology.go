package topology

import (
	"encoding/json"
	"io"
	"os"
)

// Parse a subset of the component fields
type Component struct {
	Type       string `json:"type"`
	ID         int    `json:"id"`
	Interfaces []int  `json:"interfaces"`
}

// Parse a subset of the SimBricks system interfaces
type Iface struct {
	Type      string `json:"type"`
	ID        int    `json:"id"`
	Component int    `json:"component"`
	Channel   int    `json:"channel"`
}

type Channel struct {
	Type    string `json:"type"`
	ID      int    `json:"id"`
	Latency int    `json:"latency"`
	IfaceA  int    `json:"interface_a"`
	IfaceB  int    `json:"interface_b"`
}

// Parse a subset of the SimBricks system fields
type System struct {
	Name       string      `json:"name"`
	Components []Component `json:"all_components"`
	Interfaces []Iface     `json:"interfaces"`
	Channels   []Channel   `json:"channels"`
}

// Parse a subset of the SimBricks simulator fields
type Simulator struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Components []int  `json:"components"`
}

// Parse a subset of the SimBricks simulation fields
type Simulation struct {
	Name       string      `json:"name"`
	Simulators []Simulator `json:"sim_list"`
}

type Topology struct {
	Sys System     `json:"system"`
	Sim Simulation `json:"simulation"`
}

func ParseTopology(filename string) (*Topology, error) {
	var topo Topology

	topo_file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer topo_file.Close()

	bytes, err := io.ReadAll(topo_file)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(bytes), &topo)
	if err != nil {
		return nil, err
	}

	return &topo, nil
}
