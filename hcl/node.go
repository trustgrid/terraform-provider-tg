package hcl

import "github.com/trustgrid/terraform-provider-tg/tg"

type Node struct {
	UID     string `tf:"node_id"`
	Enabled bool   `tf:"enabled"`
}

func (h Node) ToTG() tg.NodeState {
	state := "ACTIVE"
	if !h.Enabled {
		state = "INACTIVE"
	}
	return tg.NodeState{
		State: state,
	}
}

func (h Node) UpdateFromTG(a tg.NodeState) HCL[tg.NodeState] {
	return Node{
		UID:     h.UID,
		Enabled: a.State == "ACTIVE",
	}
}
