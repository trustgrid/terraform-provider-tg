package hcl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func TestContainerState_ToTG(t *testing.T) {
	tests := []struct {
		name     string
		hclState ContainerState
		want     tg.ContainerState
	}{
		{
			name: "converts all fields with node ID",
			hclState: ContainerState{
				NodeID:      "node-123",
				ClusterFQDN: "",
				ContainerID: "container-456",
				Enabled:     true,
			},
			want: tg.ContainerState{
				NodeID:      "node-123",
				ClusterFQDN: "",
				ContainerID: "container-456",
				Enabled:     true,
			},
		},
		{
			name: "converts all fields with cluster FQDN",
			hclState: ContainerState{
				NodeID:      "",
				ClusterFQDN: "cluster.example.com",
				ContainerID: "container-789",
				Enabled:     false,
			},
			want: tg.ContainerState{
				NodeID:      "",
				ClusterFQDN: "cluster.example.com",
				ContainerID: "container-789",
				Enabled:     false,
			},
		},
		{
			name: "handles enabled=false",
			hclState: ContainerState{
				NodeID:      "node-abc",
				ClusterFQDN: "",
				ContainerID: "my-container",
				Enabled:     false,
			},
			want: tg.ContainerState{
				NodeID:      "node-abc",
				ClusterFQDN: "",
				ContainerID: "my-container",
				Enabled:     false,
			},
		},
		{
			name:     "handles zero value struct",
			hclState: ContainerState{},
			want:     tg.ContainerState{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.hclState.ToTG()

			assert.Equal(t, tt.want.NodeID, got.NodeID, "NodeID mismatch")
			assert.Equal(t, tt.want.ClusterFQDN, got.ClusterFQDN, "ClusterFQDN mismatch")
			assert.Equal(t, tt.want.ContainerID, got.ContainerID, "ContainerID mismatch")
			assert.Equal(t, tt.want.Enabled, got.Enabled, "Enabled mismatch")
		})
	}
}

func TestContainerState_UpdateFromTG(t *testing.T) {
	tests := []struct {
		name        string
		existing    ContainerState
		tgState     tg.ContainerState
		wantNodeID  string
		wantCluster string
		wantCID     string
		wantEnabled bool
	}{
		{
			name: "updates enabled from API response",
			existing: ContainerState{
				NodeID:      "node-123",
				ClusterFQDN: "",
				ContainerID: "container-456",
				Enabled:     false,
			},
			tgState: tg.ContainerState{
				NodeID:      "ignored-node",
				ClusterFQDN: "ignored-cluster",
				ContainerID: "ignored-container",
				Enabled:     true,
			},
			wantNodeID:  "node-123",
			wantCluster: "",
			wantCID:     "container-456",
			wantEnabled: true,
		},
		{
			name: "preserves HCL values for non-API fields",
			existing: ContainerState{
				NodeID:      "my-node",
				ClusterFQDN: "my-cluster.fqdn",
				ContainerID: "my-container",
				Enabled:     true,
			},
			tgState: tg.ContainerState{
				Enabled: false,
			},
			wantNodeID:  "my-node",
			wantCluster: "my-cluster.fqdn",
			wantCID:     "my-container",
			wantEnabled: false,
		},
		{
			name:     "handles zero value existing state",
			existing: ContainerState{},
			tgState: tg.ContainerState{
				Enabled: true,
			},
			wantNodeID:  "",
			wantCluster: "",
			wantCID:     "",
			wantEnabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.existing.UpdateFromTG(tt.tgState)

			updated, ok := result.(ContainerState)
			require.True(t, ok, "UpdateFromTG should return ContainerState type")

			assert.Equal(t, tt.wantNodeID, updated.NodeID, "NodeID should be preserved from HCL")
			assert.Equal(t, tt.wantCluster, updated.ClusterFQDN, "ClusterFQDN should be preserved from HCL")
			assert.Equal(t, tt.wantCID, updated.ContainerID, "ContainerID should be preserved from HCL")
			assert.Equal(t, tt.wantEnabled, updated.Enabled, "Enabled should come from API response")
		})
	}
}

func TestContainerState_RoundTrip(t *testing.T) {
	// Test that converting HCL -> TG -> HCL preserves the data correctly
	original := ContainerState{
		NodeID:      "node-roundtrip",
		ClusterFQDN: "",
		ContainerID: "container-roundtrip",
		Enabled:     true,
	}

	// Convert to TG
	tgState := original.ToTG()

	// Convert back to HCL
	result := original.UpdateFromTG(tgState)
	roundTripped, ok := result.(ContainerState)
	require.True(t, ok, "UpdateFromTG should return ContainerState type")

	assert.Equal(t, original.NodeID, roundTripped.NodeID, "NodeID should survive round trip")
	assert.Equal(t, original.ClusterFQDN, roundTripped.ClusterFQDN, "ClusterFQDN should survive round trip")
	assert.Equal(t, original.ContainerID, roundTripped.ContainerID, "ContainerID should survive round trip")
	assert.Equal(t, original.Enabled, roundTripped.Enabled, "Enabled should survive round trip")
}
