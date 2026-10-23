package acctests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/trustgrid/terraform-provider-tg/provider"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

// TestAccNetworkConfig_WireGuard uses Terraform and a local HTTPS API fixture;
// it requires no Portal credentials and never changes a real node.
func TestAccNetworkConfig_WireGuard(t *testing.T) {
	const nodeID = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
	// Synthetic fixture keys, never used by a real peer.
	const publicKey = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	const presharedKey = "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE="
	const rotatedKey = "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI="

	var mu sync.Mutex
	var network map[string]any
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/node/"+nodeID+"/config/network":
			if err := json.NewDecoder(r.Body).Decode(&network); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{}))
		case r.Method == http.MethodGet && r.URL.Path == "/node/"+nodeID:
			assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{
				"uid": nodeID, "config": map[string]any{"network": network},
			}))
		default:
			t.Errorf("unexpected fixture request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	previousClient := http.DefaultClient
	http.DefaultClient = server.Client()
	t.Cleanup(func() { http.DefaultClient = previousClient })

	p := provider.New("test")()
	p.ConfigureContextFunc = func(context.Context, *schema.ResourceData) (any, diag.Diagnostics) {
		return &tg.Client{APIHost: strings.TrimPrefix(server.URL, "https://")}, nil
	}

	var steps []resource.TestStep
	for _, key := range []string{presharedKey, rotatedKey, ""} {
		pskConfig := ""
		if key != "" {
			pskConfig = fmt.Sprintf("preshared_key = %q", key)
		}
		config := fmt.Sprintf(`
resource "tg_network_config" "test" {
  node_id = %q
  tunnel {
    type = "wireguard"
    name = "wg1"
    enabled = true
    ip = "192.0.2.1/24"
    public_key = %q
    %s
    mtu = 1430
    vrf = "vpn"
    description = "WireGuard fixture"
  }
}
data "tg_network_config" "test" {
  node_id = tg_network_config.test.id
  depends_on = [tg_network_config.test]
}
`, nodeID, publicKey, pskConfig)
		steps = append(steps, resource.TestStep{
			Config: config,
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectSensitiveValue("tg_network_config.test", tfjsonpath.New("tunnel").AtSliceIndex(0).AtMapKey("preshared_key")),
				},
			},
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("tg_network_config.test", "tunnel.0.public_key", publicKey),
				resource.TestCheckResourceAttr("tg_network_config.test", "tunnel.0.preshared_key", key),
				resource.TestCheckResourceAttr("data.tg_network_config.test", "tunnel.0.public_key", publicKey),
				resource.TestCheckResourceAttr("data.tg_network_config.test", "tunnel.0.preshared_key", key),
			),
		})
		steps = append(steps, resource.TestStep{
			Config: config, PlanOnly: true, ExpectNonEmptyPlan: false,
			PreConfig: func() {
				mu.Lock()
				defer mu.Unlock()
				tunnels, ok := network["tunnels"].([]any)
				require.True(t, ok)
				require.Len(t, tunnels, 1)
				tunnel, ok := tunnels[0].(map[string]any)
				require.True(t, ok)
				for field, want := range map[string]any{
					"type": "wireguard", "name": "wg1", "enabled": true,
					"ip": "192.0.2.1/24", "publicKey": publicKey,
					"mtu": float64(1430), "vrf": "vpn", "description": "WireGuard fixture",
				} {
					assert.Equal(t, want, tunnel[field], field)
				}
				if key == "" {
					assert.NotContains(t, tunnel, "presharedKey")
				} else {
					assert.Equal(t, key, tunnel["presharedKey"])
				}
				assert.NotContains(t, tunnel, "public_key")
				assert.NotContains(t, tunnel, "preshared_key")
				assert.NotContains(t, tunnel, "psk")
			},
		})
	}
	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  map[string]*schema.Provider{"tg": p},
		Steps:      steps,
	})
}
