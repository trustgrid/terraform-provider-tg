package acctests

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/trustgrid/terraform-provider-tg/provider"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

// Test creates its own cluster (with a random-suffix name to avoid cluster-
// name V2 state retention on the backend), upgrades it to V2, and exercises
// tg_cluster_connector. TestCase auto-destroy at end wipes the cluster and
// every connector inside it.

func TestAccClusterConnectorV2_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())
	p := provider.New("test")()
	clusterName := "tf-test-v2-cc-" + acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"tg": p},
		Steps: []resource.TestStep{
			{
				Config: clusterConnectorV2Config(clusterName, 100),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_cluster_connector.test", "connector_id"),
					resource.TestCheckResourceAttr("tg_cluster_connector.test", "port", "9090"),
					resource.TestCheckResourceAttr("tg_cluster_connector.test", "protocol", "tcp"),
					resource.TestCheckResourceAttr("tg_cluster_connector.test", "rate_limit", "100"),
					checkClusterConnectorAPISide(p, clusterName, 100),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_cluster_connector.test", tfjsonpath.New("connector_id")),
				},
			},
			{
				Config: clusterConnectorV2Config(clusterName, 500),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_cluster_connector.test", "rate_limit", "500"),
					checkClusterConnectorAPISide(p, clusterName, 500),
				),
			},
		},
	})
}

func clusterConnectorV2Config(clusterName string, rateLimit int) string {
	return fmt.Sprintf(`
resource "tg_cluster" "test" {
  name = %q
}

resource "tg_cluster_connectors_v2_upgrade" "test" {
  cluster_fqdn = tg_cluster.test.fqdn
}

resource "tg_cluster_connector" "test" {
  cluster_fqdn = tg_cluster.test.fqdn
  depends_on   = [tg_cluster_connectors_v2_upgrade.test]
  node         = "local"
  service      = "127.0.0.1:9091"
  port         = 9090
  protocol     = "tcp"
  description  = "tf-test-v2-cluster-conn"
  enabled      = true
  nic          = "any"
  rate_limit   = %d
}
`, clusterName, rateLimit)
}

func checkClusterConnectorAPISide(p *schema.Provider, clusterName string, wantRateLimit int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := p.Meta().(*tg.Client)
		fqdn := clusterName + "." + client.Domain

		rs, ok := s.RootModule().Resources["tg_cluster_connector.test"]
		if !ok {
			return fmt.Errorf("tg_cluster_connector.test not found in state")
		}
		connectorID := rs.Primary.ID

		var cluster tg.Cluster
		if err := client.Get(context.Background(), fmt.Sprintf("/cluster/%s", fqdn), &cluster); err != nil {
			return fmt.Errorf("error getting cluster: %w", err)
		}
		if cluster.Config.Connectors == nil {
			return fmt.Errorf("cluster has no connectors config")
		}
		var found *tg.Connector
		for i := range cluster.Config.Connectors.Connectors {
			if cluster.Config.Connectors.Connectors[i].ID == connectorID {
				found = &cluster.Config.Connectors.Connectors[i]
				break
			}
		}
		if found == nil {
			return fmt.Errorf("connector %s not found in cluster config", connectorID)
		}
		if found.RateLimit != wantRateLimit {
			return fmt.Errorf("expected rate_limit %d, got %d", wantRateLimit, found.RateLimit)
		}
		return nil
	}
}
