package symphony

import (
	"github.com/vaastav/columbo_go/components"
)

type HostNicPair struct {
	Host components.Component
	NIC  components.Component
}

type EthPair struct {
	A components.Component
	B components.Component
}

func HostNicPairs(sim *Simulation) ([]HostNicPair, error) {
	var pairs []HostNicPair
	for _, channel := range sim.Channels {
		if channel.ChanData.Type == "PCIeChannel" {
			pair := HostNicPair{}
			ifacea := channel.IfaceA
			ifaceb := channel.IfaceB

			if ifacea.Type == "PCIeHostInterface" {
				pair.Host = sim.Components[ifacea.Component]
				pair.NIC = sim.Components[ifaceb.Component]
			} else {
				pair.Host = sim.Components[ifaceb.Component]
				pair.NIC = sim.Components[ifacea.Component]
			}
			pairs = append(pairs, pair)
		}
	}
	return pairs, nil
}

func Switches(sim *Simulation) ([]components.Component, error) {
	var switches []components.Component
	for _, component := range sim.Components {
		if _, ok := component.(*components.Switch); ok {
			switches = append(switches, component)
		}
	}
	return switches, nil
}

func Hosts(sim *Simulation) ([]components.Component, error) {
	var hosts []components.Component
	for _, component := range sim.Components {
		if _, ok := component.(*components.Host); ok {
			hosts = append(hosts, component)
		}
	}
	return hosts, nil
}

func NICs(sim *Simulation) ([]components.Component, error) {
	var nics []components.Component
	for _, component := range sim.Components {
		if _, ok := component.(*components.NIC); ok {
			nics = append(nics, component)
		}
	}
	return nics, nil
}

func EthPairs(sim *Simulation) ([]EthPair, error) {
	var pairs []EthPair
	for _, channel := range sim.Channels {
		if channel.ChanData.Type == "EthChannel" {
			pair := EthPair{}
			ifacea := channel.IfaceA
			ifaceb := channel.IfaceB

			pair.A = sim.Components[ifacea.Component]
			pair.B = sim.Components[ifaceb.Component]
			pairs = append(pairs, pair)
		}
	}
	return pairs, nil
}

func SwitchPairs(sim *Simulation) ([]EthPair, error) {
	pairs, err := EthPairs(sim)
	if err != nil {
		return pairs, err
	}
	// Filter the pairs
	var res []EthPair
	for _, pair := range pairs {
		_, ok1 := pair.A.(*components.Switch)
		_, ok2 := pair.B.(*components.Switch)
		if ok1 && ok2 {
			res = append(res, pair)
		}
	}
	return res, nil
}

func NicSwitchPairs(sim *Simulation) ([]EthPair, error) {
	pairs, err := EthPairs(sim)
	if err != nil {
		return pairs, err
	}
	// Filter the pairs
	var res []EthPair
	for _, pair := range pairs {
		if _, ok := pair.A.(*components.Switch); ok {
			if _, ok2 := pair.B.(*components.NIC); ok2 {
				res = append(res, pair)
			}
		} else {
			if _, ok2 := pair.B.(*components.Switch); ok2 {
				res = append(res, pair)
			}
		}
	}
	return pairs, nil
}
