package resource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

// ifaceEndpoint returns (id, isCluster) from resource data.
// node_id takes priority; falls back to cluster_fqdn.
func ifaceEndpoint(d *schema.ResourceData) (string, bool) {
	nodeID, ok := d.GetOk("node_id")
	if ok {
		if id, ok := nodeID.(string); ok && id != "" {
			return id, false
		}
	}
	id, _ := d.Get("cluster_fqdn").(string)
	return id, true
}

// getNetworkConfig fetches the current full NetworkConfig for a node or cluster.
func getNetworkConfig(ctx context.Context, tgc *tg.Client, id string, isCluster bool) (tg.NetworkConfig, error) {
	if isCluster {
		n := tg.Cluster{}
		if err := tgc.Get(ctx, "/cluster/"+id, &n); err != nil {
			return tg.NetworkConfig{}, err
		}
		if n.Config.Network != nil {
			return *n.Config.Network, nil
		}
		return tg.NetworkConfig{}, nil
	}
	n := tg.Node{}
	if err := tgc.Get(ctx, "/node/"+id, &n); err != nil {
		return tg.NetworkConfig{}, err
	}
	return n.Config.Network, nil
}

// putNetworkConfig writes the full NetworkConfig back to the API.
func putNetworkConfig(ctx context.Context, tgc *tg.Client, id string, isCluster bool, nc tg.NetworkConfig) error {
	if isCluster {
		_, err := tgc.Put(ctx, fmt.Sprintf("/cluster/%s/config/network", id), &nc)
		return err
	}
	_, err := tgc.Put(ctx, fmt.Sprintf("/node/%s/config/network", id), &nc)
	return err
}

// encodeIfaceID builds "{endpoint}:{nic}".
func encodeIfaceID(endpoint, nic string) string {
	return endpoint + ":" + nic
}

// encodeIfaceRouteID builds "{endpoint}:{nic}:{route}".
func encodeIfaceRouteID(endpoint, nic, route string) string {
	return endpoint + ":" + nic + ":" + route
}

// encodeIfaceVLANID builds "{endpoint}:{nic}:{vlan_id}".
func encodeIfaceVLANID(endpoint, nic string, vlanID int) string {
	return fmt.Sprintf("%s:%s:%d", endpoint, nic, vlanID)
}
