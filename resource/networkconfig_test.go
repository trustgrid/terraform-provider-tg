package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trustgrid/terraform-provider-tg/hcl"
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
