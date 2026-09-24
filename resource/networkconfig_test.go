package resource

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func TestValidateNetworkConfigInterfaces(t *testing.T) {
	tests := []struct {
		name       string
		interfaces []hcl.NetworkInterface
		isCluster  bool
		err        string
	}{
		{
			name:       "node allows dhcp",
			interfaces: []hcl.NetworkInterface{{NIC: "ens192", DHCP: true}},
			isCluster:  false,
		},
		{
			name:       "cluster allows interface without dhcp",
			interfaces: []hcl.NetworkInterface{{NIC: "ens192"}},
			isCluster:  true,
		},
		{
			name:       "cluster rejects dhcp",
			interfaces: []hcl.NetworkInterface{{NIC: "ens192", DHCP: true}},
			isCluster:  true,
			err:        `interface "ens192" cannot set dhcp = true when cluster_fqdn is set`,
		},
		{
			name:       "cluster rejects dhcp without nic",
			interfaces: []hcl.NetworkInterface{{DHCP: true}},
			isCluster:  true,
			err:        `interface "index 0" cannot set dhcp = true when cluster_fqdn is set`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNetworkConfigInterfaces(tt.interfaces, tt.isCluster)
			if tt.err == "" {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			assert.ErrorContains(t, err, tt.err)
		})
	}
}

func TestNetworkConfig_VRFRuleRoundTrip(t *testing.T) {
	rules := []any{
		map[string]any{
			"protocol": "udp",
			"line":     10,
			"action":   "forward",
			"source":   "0.0.0.0/0",
			"dest":     "0.0.0.0/0",
			"in":       "ens224",
			"iface":    "ens192",
			"snat":     true,
			"ports":    "3478-3479",
		},
		map[string]any{
			"protocol": "tcp",
			"line":     20,
			"action":   "dnat",
			"source":   "0.0.0.0/0",
			"dest":     "10.0.0.1/32",
			"ports":    "8443",
			"dnat":     "10.20.0.5:443",
		},
	}
	raw := map[string]any{
		"node_id": "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48",
		"vrf": []any{map[string]any{
			"name": "blue",
			"rule": rules,
		}},
	}

	d := schema.TestResourceDataRaw(t, NetworkConfig().Schema, raw)
	tf, err := hcl.DecodeResourceData[hcl.NetworkConfig](d)
	require.NoError(t, err)

	body, err := json.Marshal(tf.ToTG())
	require.NoError(t, err)

	var wire struct {
		VRFs []struct {
			Rules []map[string]any `json:"rules"`
		} `json:"vrfs"`
	}
	require.NoError(t, json.Unmarshal(body, &wire))
	require.Len(t, wire.VRFs, 1)
	require.Len(t, wire.VRFs[0].Rules, 2)

	forward := wire.VRFs[0].Rules[0]
	assert.Equal(t, "ens224", forward["in"])
	assert.Equal(t, "ens192", forward["iface"])
	assert.Equal(t, true, forward["snat"])
	assert.Equal(t, "3478-3479", forward["ports"])
	assert.NotContains(t, forward, "dnat")

	dnat := wire.VRFs[0].Rules[1]
	assert.Equal(t, "10.20.0.5:443", dnat["dnat"])
	assert.Equal(t, "8443", dnat["ports"])
	assert.NotContains(t, dnat, "snat")
	assert.NotContains(t, dnat, "iface")
	assert.NotContains(t, dnat, "in")

	var fromAPI tg.NetworkConfig
	require.NoError(t, json.Unmarshal(body, &fromAPI))
	readBack := hcl.NetworkConfig{}
	readBack.UpdateFromTG(fromAPI)

	out := schema.TestResourceDataRaw(t, NetworkConfig().Schema, map[string]any{})
	require.NoError(t, hcl.EncodeResourceData(readBack, out))

	for i, r := range rules {
		for k, want := range r.(map[string]any) {
			assert.Equal(t, want, out.Get(fmt.Sprintf("vrf.0.rule.%d.%s", i, k)), "rule %d field %s", i, k)
		}
	}
	assert.Equal(t, false, out.Get("vrf.0.rule.1.snat"))
	assert.Equal(t, "", out.Get("vrf.0.rule.1.iface"))
}
