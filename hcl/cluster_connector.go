package hcl

import (
	"github.com/trustgrid/terraform-provider-tg/tg"
)

// ClusterConnector is the HCL representation of a V2 cluster connector.
// Writes go through /v2/cluster/{cluster_fqdn}/config/connectors[/{connector_id}].
type ClusterConnector struct {
	ConnectorID string `tf:"connector_id"`
	ClusterFQDN string `tf:"cluster_fqdn"`
	Node        string `tf:"node"`
	Service     string `tf:"service"`
	Port        int    `tf:"port"`
	Protocol    string `tf:"protocol"`
	Description string `tf:"description"`
	Enabled     bool   `tf:"enabled"`
	RateLimit   int    `tf:"rate_limit"`
	NIC         string `tf:"nic"`
}

func (c ClusterConnector) ToTG() tg.Connector {
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

func (c ClusterConnector) UpdateFromTG(conn tg.Connector) HCL[tg.Connector] {
	return ClusterConnector{
		ConnectorID: conn.ID,
		ClusterFQDN: c.ClusterFQDN,
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
