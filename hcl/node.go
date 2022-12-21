package hcl

import "github.com/trustgrid/terraform-provider-tg/tg"

type Node struct {
	UID     string `tf:"node_id"`
	Enabled bool   `tf:"enabled"`
}

func (h *Node) ResourceURL(ID string) string {
	return h.URL() + "/" + ID
}

func (h *Node) URL() string {
	return "/node"
}

func (h *Node) ToTG() tg.NodeState {
	state := "enabled"
	if !h.Enabled {
		state = "disabled"
	}
	return tg.NodeState{
		State: state,
	}
}

func (h *Node) UpdateFromTG(a tg.NodeState) {
	h.Enabled = a.State == "enabled"
}
