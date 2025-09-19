package symphony

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/vaastav/columbo_go/components"
	"github.com/vaastav/columbo_go/parser"
	"github.com/vaastav/columbo_go/simulators/gem5"
	"github.com/vaastav/columbo_go/simulators/netswitch"
	"github.com/vaastav/columbo_go/simulators/nicbm"
	"github.com/vaastav/columbo_go/topology"
)

type SimInstance struct {
	ID         int64
	P          parser.Parser
	Components []components.Component
}

func NewSimInstance(ID int64, P parser.Parser, comps []components.Component) (*SimInstance, error) {
	return &SimInstance{ID: ID, P: P, Components: comps}, nil
}

func (s *SimInstance) Process(ctx context.Context, line string) error {
	// Parse the event
	e, err := s.P.ParseEvent(line)
	if err != nil {
		return err
	}
	// Push events to the components
	for _, c := range s.Components {
		err = c.HandleEvent(*e)
		if err != nil {
			return err
		}
	}
	return nil
}
func CreateSimInstanceFromTopology(ctx context.Context, topo *topology.Topology, BUFFER_SIZE int) (map[string]*SimInstance, error) {
	instances := make(map[string]*SimInstance)
	component_map := make(map[int]components.Component)
	for _, cmp := range topo.Sys.Components {
		var comp components.Component
		var err error
		name := fmt.Sprintf("comp-%d", cmp.ID)
		switch cmp.Type {
		case "I40ELinuxHost":
			comp, err = components.NewHost(ctx, name, cmp.ID, BUFFER_SIZE)
		case "IntelI40eNIC":
			comp, err = components.NewNIC(ctx, name, cmp.ID, BUFFER_SIZE)
		case "EthSwitch":
			comp, err = components.NewSwitch(ctx, name, cmp.ID, BUFFER_SIZE)
		default:
			err = errors.New("Unsupported component type " + cmp.Type)
		}
		if err != nil {
			return nil, err
		}
		component_map[cmp.ID] = comp
	}
	for _, sim := range topo.Sim.Simulators {
		var components []components.Component
		for _, cmp_id := range sim.Components {
			if c, ok := component_map[cmp_id]; ok {
				components = append(components, c)
			} else {
				log.Fatalf("Failed to find component with id %d\n", cmp_id)
			}
		}
		var p parser.Parser
		var err error
		switch sim.Type {
		case "Gem5Sim":
			p, err = gem5.NewGem5Parser(ctx, sim.ID, sim.Name)
		case "I40eNicSim":
			p, err = nicbm.NewNicBMParser(ctx, sim.ID, sim.Name)
		case "SwitchNet":
			p, err = netswitch.NewNetSwitchParser(ctx, sim.ID, sim.Name)
		default:
			return nil, errors.New("Unsupported parser type " + sim.Type)
		}
		if err != nil {
			return nil, err
		}
		instance, err := NewSimInstance(sim.ID, p, components)
		if err != nil {
			return nil, err
		}
		instances[sim.Name] = instance
	}

	return instances, nil
}
