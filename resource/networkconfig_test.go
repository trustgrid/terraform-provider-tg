package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateNetworkConfigInterfaces(t *testing.T) {
	tests := []struct {
		name       string
		interfaces []any
		isCluster  bool
		err        string
	}{
		{
			name: "node allows dhcp",
			interfaces: []any{map[string]any{
				"nic":  "ens192",
				"dhcp": true,
			}},
			isCluster: false,
		},
		{
			name: "cluster allows interface without dhcp",
			interfaces: []any{map[string]any{
				"nic": "ens192",
			}},
			isCluster: true,
		},
		{
			name: "cluster rejects dhcp",
			interfaces: []any{map[string]any{
				"nic":  "ens192",
				"dhcp": true,
			}},
			isCluster: true,
			err:       `interface "ens192" cannot set dhcp = true when cluster_fqdn is set`,
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
