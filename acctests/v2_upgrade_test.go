package acctests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/trustgrid/terraform-provider-tg/provider"
)

// TestAccClusterServicesV2Upgrade_Idempotent verifies that applying the
// upgrade resource twice against the same cluster is a no-op the second time —
// the API returns 422 when re-upgrading an already-V2 cluster, and the
// provider silently treats that as success.
//
// Uses a fresh cluster per test (auto-destroyed by TestCase teardown), so the
// pattern works regardless of CI state.
func TestAccClusterServicesV2Upgrade_Idempotent(t *testing.T) {
	clusterName := "tf-test-v2-sup-" + acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"tg": provider.New("test")()},
		Steps: []resource.TestStep{
			{
				Config: clusterServicesV2UpgradeConfig(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_cluster_services_v2_upgrade.test", "id"),
				),
			},
			// Re-apply same config — must be no-op against an already-V2 cluster.
			{
				Config:             clusterServicesV2UpgradeConfig(clusterName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccClusterConnectorsV2Upgrade_Idempotent(t *testing.T) {
	clusterName := "tf-test-v2-cup-" + acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"tg": provider.New("test")()},
		Steps: []resource.TestStep{
			{
				Config: clusterConnectorsV2UpgradeConfig(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_cluster_connectors_v2_upgrade.test", "id"),
				),
			},
			{
				Config:             clusterConnectorsV2UpgradeConfig(clusterName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func clusterServicesV2UpgradeConfig(clusterName string) string {
	return fmt.Sprintf(`
resource "tg_cluster" "test" {
  name = %q
}

resource "tg_cluster_services_v2_upgrade" "test" {
  cluster_fqdn = tg_cluster.test.fqdn
}
`, clusterName)
}

func clusterConnectorsV2UpgradeConfig(clusterName string) string {
	return fmt.Sprintf(`
resource "tg_cluster" "test" {
  name = %q
}

resource "tg_cluster_connectors_v2_upgrade" "test" {
  cluster_fqdn = tg_cluster.test.fqdn
}
`, clusterName)
}
