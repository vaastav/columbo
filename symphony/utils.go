package symphony

import (
	"log"

	"github.com/vaastav/columbo_go/components"
)

type HostNicPair struct {
	Host components.Component
	NIC  components.Component
}

func HostNicPairs(sim *Simulation) ([]HostNicPair, error) {
	var pairs []HostNicPair
	for _, channel := range sim.Channels {
		if channel.ChanData.Type == "PCIeChannel" {
			log.Println("Inside pciechannel")
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
