package hcl

import (
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type Connector struct {
	NodeID      string `tf:"node_id"`
	ClusterFQDN string `tf:"cluster_fqdn"`
	Node        string `tf:"node"`
	Port        int    `tf:"port"`
	Protocol    string `tf:"protocol"`
	Service     string `tf:"service"`
	Description string `tf:"description"`
	RateLimit   int    `tf:"rate_limit,omitempty"`
}

func (c *Connector) UpdateFromTG(conn tg.Connector) {
	c.Node = conn.Node
	c.Service = conn.Service
	c.Port = conn.Port
	c.Protocol = conn.Protocol
	c.Description = conn.Description
	c.RateLimit = conn.RateLimit
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
		RateLimit:   c.RateLimit,
	}
}
