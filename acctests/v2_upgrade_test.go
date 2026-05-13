package acctests

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/trustgrid/terraform-provider-tg/provider"
)

// V1→V2 upgrade is one-way and irreversible per the API. These tests assume
// the testClusterFQDN cluster and testNodeID node have ALREADY been upgraded
// to V2. We test the resource's idempotency (state present, no error on
// subsequent applies) and the no-op destroy behavior. We do NOT trigger a
// fresh V1→V2 upgrade here because that's not reversible.

func TestAccClusterServicesV2Upgrade_Idempotent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"tg": provider.New("test")()},
		Steps: []resource.TestStep{
			{
				Config: clusterServicesV2UpgradeConfig(testClusterFQDN),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_cluster_services_v2_upgrade.test", "cluster_fqdn", testClusterFQDN),
					resource.TestCheckResourceAttr("tg_cluster_services_v2_upgrade.test", "id", testClusterFQDN),
				),
			},
			// Re-apply same config: must be no-op.
			{
				Config:             clusterServicesV2UpgradeConfig(testClusterFQDN),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccClusterConnectorsV2Upgrade_Idempotent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"tg": provider.New("test")()},
		Steps: []resource.TestStep{
			{
				Config: clusterConnectorsV2UpgradeConfig(testClusterFQDN),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_cluster_connectors_v2_upgrade.test", "cluster_fqdn", testClusterFQDN),
					resource.TestCheckResourceAttr("tg_cluster_connectors_v2_upgrade.test", "id", testClusterFQDN),
				),
			},
			{
				Config:             clusterConnectorsV2UpgradeConfig(testClusterFQDN),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func clusterServicesV2UpgradeConfig(fqdn string) string {
	return `
resource "tg_cluster_services_v2_upgrade" "test" {
  cluster_fqdn = "` + fqdn + `"
}
`
}

func clusterConnectorsV2UpgradeConfig(fqdn string) string {
	return `
resource "tg_cluster_connectors_v2_upgrade" "test" {
  cluster_fqdn = "` + fqdn + `"
}
`
}
