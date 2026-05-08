package resource

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func intPointer(v int) *int {
	return &v
}

func TestValidateVNetRouteMonitors(t *testing.T) {
	tests := []struct {
		name     string
		monitors []any
		err      string
	}{
		{
			name: "tcp valid",
			monitors: []any{map[string]any{
				"name":        "tcp-probe",
				"enabled":     true,
				"protocol":    "tcp",
				"dest":        "10.100.0.10",
				"port":        443,
				"interval":    5,
				"count":       3,
				"max_latency": 500,
			}},
		},
		{
			name: "icmp valid",
			monitors: []any{map[string]any{
				"name":     "icmp-probe",
				"enabled":  true,
				"protocol": "icmp",
				"dest":     "10.100.0.10",
				"interval": 5,
				"count":    3,
			}},
		},
		{
			name: "enabled must be true",
			monitors: []any{map[string]any{
				"name":     "icmp-probe",
				"enabled":  false,
				"protocol": "icmp",
				"dest":     "10.100.0.10",
				"interval": 5,
				"count":    3,
			}},
			err: "enabled must be true",
		},
		{
			name: "tcp requires port",
			monitors: []any{map[string]any{
				"name":     "tcp-probe",
				"enabled":  true,
				"protocol": "tcp",
				"dest":     "10.100.0.10",
				"interval": 5,
				"count":    3,
			}},
			err: "port is required when protocol is tcp",
		},
		{
			name: "icmp forbids port",
			monitors: []any{map[string]any{
				"name":     "icmp-probe",
				"enabled":  true,
				"protocol": "icmp",
				"dest":     "10.100.0.10",
				"port":     443,
				"interval": 5,
				"count":    3,
			}},
			err: "port must not be set when protocol is icmp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVNetRouteMonitors(tt.monitors)
			if tt.err == "" {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			assert.ErrorContains(t, err, tt.err)
		})
	}
}

func TestVNetRouteMonitorEncodeDecode(t *testing.T) {
	res := VNetRoute()
	d := (&schema.Resource{Schema: res.Schema}).TestResourceData()

	err := d.Set("network", "test-network")
	require.NoError(t, err)
	err = d.Set("dest", "edge-node")
	require.NoError(t, err)
	err = d.Set("network_cidr", "10.10.10.14/32")
	require.NoError(t, err)
	err = d.Set("metric", 1)
	require.NoError(t, err)
	err = d.Set("description", "my edge node route")
	require.NoError(t, err)
	err = d.Set("monitor", []map[string]any{{
		"name":        "tcp-probe",
		"enabled":     true,
		"protocol":    "tcp",
		"dest":        "10.100.0.10",
		"port":        443,
		"interval":    5,
		"count":       3,
		"max_latency": 500,
	}})
	require.NoError(t, err)

	route, err := hcl.DecodeResourceData[tg.VNetRoute](d)
	require.NoError(t, err)
	require.Len(t, route.Monitors, 1)
	assert.Equal(t, tg.VNetRouteMonitor{
		Name:       "tcp-probe",
		Enabled:    true,
		Protocol:   "tcp",
		Dest:       "10.100.0.10",
		Port:       intPointer(443),
		Interval:   5,
		Count:      3,
		MaxLatency: intPointer(500),
	}, route.Monitors[0])

	route.UID = "route-uid"
	d = (&schema.Resource{Schema: res.Schema}).TestResourceData()
	err = hcl.EncodeResourceData(route, d)
	require.NoError(t, err)

	assert.Equal(t, "route-uid", d.Get("uid"))
	monitors := d.Get("monitor").([]any)
	require.Len(t, monitors, 1)
	monitor := monitors[0].(map[string]any)
	assert.Equal(t, "tcp-probe", monitor["name"])
	assert.Equal(t, true, monitor["enabled"])
	assert.Equal(t, "tcp", monitor["protocol"])
	assert.Equal(t, "10.100.0.10", monitor["dest"])
	assert.Equal(t, 443, monitor["port"])
	assert.Equal(t, 5, monitor["interval"])
	assert.Equal(t, 3, monitor["count"])
	assert.Equal(t, 500, monitor["max_latency"])
}
