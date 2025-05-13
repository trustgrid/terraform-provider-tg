package hcl

import (
	"fmt"

	"github.com/trustgrid/terraform-provider-tg/tg"
)

// VPN refers to node/cluster VPN config
// virtualnetwork refers to domain-level vnet config

type VPNDynamicRoute struct {
	UID         string `tf:"uid"`
	NodeID      string `tf:"node_id"`
	NetworkName string `tf:"network_name"`

	Description string `tf:"description"`
	Metric      int    `tf:"metric"`
	NetworkCIDR string `tf:"network_cidr"`
	Path        string `tf:"path"`
	Node        string `tf:"node"`
}

func (r VPNDynamicRoute) UpdateFromTG(route tg.VPNDynamicExportRoute) HCL[tg.VPNDynamicExportRoute] {
	return VPNDynamicRoute{
		UID:         r.UID,
		NodeID:      r.NodeID,
		NetworkName: r.NetworkName,
		Description: route.Description,
		Metric:      route.Metric,
		NetworkCIDR: route.NetworkCIDR,
		Path:        route.Path,
		Node:        route.Node,
	}
}

func (r VPNDynamicRoute) ToTG() tg.VPNDynamicExportRoute {
	return tg.VPNDynamicExportRoute{
		UID:         r.UID,
		Description: r.Description,
		Metric:      r.Metric,
		NetworkCIDR: r.NetworkCIDR,
		Path:        r.Path,
		Node:        r.Node,
	}
}

type VPNAttachment struct {
	NodeID         string `tf:"node_id"`
	ClusterFQDN    string `tf:"cluster_fqdn"`
	NetworkName    string `tf:"network"`
	IP             string `tf:"ip"`
	ValidationCIDR string `tf:"validation_cidr"`
}

func (h *VPNAttachment) ResourceURL() string {
	return h.URL() + "/" + h.NetworkName
}

func (h *VPNAttachment) URL() string {
	if h.NodeID != "" {
		return fmt.Sprintf("/v2/node/%s/vpn", h.NodeID)
	}
	return fmt.Sprintf("/v2/cluster/%s/vpn", h.ClusterFQDN)
}
