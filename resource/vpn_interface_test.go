package resource

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

// outsideNATBody decodes an HCL tg_vpn_interface config the way Create does
// and returns the JSON the provider would POST for its outside NATs.
func outsideNATBody(t *testing.T, outside map[string]any) []map[string]any {
	t.Helper()
	raw := map[string]any{
		"node_id":        "node-uid",
		"network":        "vnet",
		"interface_name": "ens192",
		"inside_nat": []any{map[string]any{
			"network_cidr": "10.0.0.0/24",
			"local_cidr":   "192.168.1.0/24",
		}},
		"outside_nat": []any{outside},
	}
	d := schema.TestResourceDataRaw(t, VPNInterface().Schema, raw)
	tf, err := hcl.DecodeResourceData[HCLVPNInterface](d)
	require.NoError(t, err)

	body, err := json.Marshal(tf.toTGVPNInterface())
	require.NoError(t, err)

	var decoded struct {
		InsideNATs  []map[string]any `json:"insideNats"`
		OutsideNATs []map[string]any `json:"outsideNats"`
	}
	require.NoError(t, json.Unmarshal(body, &decoded))
	require.Len(t, decoded.InsideNATs, 1)
	_, insideHasProxyARP := decoded.InsideNATs[0]["proxyArp"]
	assert.False(t, insideHasProxyARP, "inside NATs must not carry proxyArp")
	return decoded.OutsideNATs
}

func TestVPNInterfaceOutsideNATProxyARP(t *testing.T) {
	base := map[string]any{"network_cidr": "10.1.0.0/24", "local_cidr": "10.1.0.0/24"}

	t.Run("explicit false is sent", func(t *testing.T) {
		nat := map[string]any{"proxy_arp": false}
		for k, v := range base {
			nat[k] = v
		}
		got := outsideNATBody(t, nat)
		require.Len(t, got, 1)
		v, ok := got[0]["proxyArp"]
		require.True(t, ok, "proxyArp=false must be in the request body")
		assert.Equal(t, false, v)
	})

	t.Run("unset defaults to true", func(t *testing.T) {
		got := outsideNATBody(t, base)
		require.Len(t, got, 1)
		assert.Equal(t, true, got[0]["proxyArp"])
	})
}

func TestVPNInterfaceReadProxyARP(t *testing.T) {
	f := false
	tr := true
	cases := []struct {
		name string
		api  *bool
		want bool
	}{
		{"missing reads as the node default", nil, true},
		{"false", &f, false},
		{"true", &tr, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var h HCLVPNInterface
			h.updateFromTGVPNInterface(tg.VPNInterface{OutsideNATs: []tg.VPNInterfaceNAT{{
				NetworkCIDR: "10.1.0.0/24", LocalCIDR: "10.1.0.0/24", ProxyARP: c.api,
			}}})
			require.Len(t, h.OutsideNATs, 1)
			assert.Equal(t, c.want, h.OutsideNATs[0].ProxyARP)
		})
	}
}
