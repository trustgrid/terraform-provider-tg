package hcl

import (
	"github.com/trustgrid/terraform-provider-tg/tg"
)

// NodeConnector is the HCL representation of a V2 node connector.
// Writes go through /v2/node/{node_id}/config/connectors[/{connector_id}].
type NodeConnector struct {
	ConnectorID string `tf:"connector_id"`
	NodeID      string `tf:"node_id"`
	Node        string `tf:"node"`
	Service     string `tf:"service"`
	Port        int    `tf:"port"`
	Protocol    string `tf:"protocol"`
	Description string `tf:"description"`
	Enabled     bool   `tf:"enabled"`
	RateLimit   int    `tf:"rate_limit"`
	NIC         string `tf:"nic"`
}

func (c NodeConnector) ToTG() tg.Connector {
	return tg.Connector{
		ID:          c.ConnectorID,
		Node:        c.Node,
		Service:     c.Service,
		Port:        c.Port,
		Protocol:    c.Protocol,
		Description: c.Description,
		Enabled:     c.Enabled,
		RateLimit:   c.RateLimit,
		NIC:         c.NIC,
	}
}

func (c NodeConnector) UpdateFromTG(conn tg.Connector) HCL[tg.Connector] {
	return NodeConnector{
		ConnectorID: conn.ID,
		NodeID:      c.NodeID,
		Node:        conn.Node,
		Service:     conn.Service,
		Port:        conn.Port,
		Protocol:    conn.Protocol,
		Description: conn.Description,
		Enabled:     conn.Enabled,
		RateLimit:   conn.RateLimit,
		NIC:         conn.NIC,
	}
}
