package hcl

import (
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type Connector struct {
	NodeID      string `tf:"node_id"`
	Node        string `tf:"node"`
	Port        int    `tf:"port"`
	Protocol    string `tf:"protocol"`
	Service     string `tf:"service"`
	Description string `tf:"description"`
}

func (c *Connector) UpdateFromTG(conn tg.Connector) {
	c.Node = conn.Node
	c.Service = conn.Service
	c.Port = conn.Port
	c.Protocol = conn.Protocol
	c.Description = conn.Description
}

func (c *Connector) ToTG(id string) tg.Connector {
	return tg.Connector{
		ID:          id,
		Node:        c.Node,
		Enabled:     true,
		Service:     c.Service,
		Port:        c.Port,
		Protocol:    c.Protocol,
		Description: c.Description,
	}
}
