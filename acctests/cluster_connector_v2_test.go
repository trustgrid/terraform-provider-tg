package acctests

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/trustgrid/terraform-provider-tg/provider"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

const v2ClusterTestConnectorDesc = "tf-test-v2-cluster-conn"

func init() {
	resource.AddTestSweepers("tg_cluster_connector", &resource.Sweeper{
		Name: "tg_cluster_connector",
		F: func(_ string) error {
			cp := tg.ClientParams{
				APIKey:    os.Getenv("TG_API_KEY_ID"),
				APISecret: os.Getenv("TG_API_KEY_SECRET"),
				APIHost:   os.Getenv("TG_API_HOST"),
			}
			client, err := tg.NewClient(context.Background(), cp)
			if err != nil {
				return fmt.Errorf("error creating client: %w", err)
			}

			var cluster tg.Cluster
			if err := client.Get(context.Background(), "/cluster/"+testClusterFQDN, &cluster); err != nil {
				return fmt.Errorf("error fetching cluster: %w", err)
			}
			if cluster.Config.Connectors == nil {
				return nil
			}
			for _, conn := range cluster.Config.Connectors.Connectors {
				if conn.Description != v2ClusterTestConnectorDesc {
					continue
				}
				url := fmt.Sprintf("/v2/cluster/%s/config/connectors/%s", testClusterFQDN, conn.ID)
				if err := client.Delete(context.Background(), url, nil); err != nil {
					return fmt.Errorf("error deleting cluster connector %s: %w", conn.ID, err)
				}
			}
			return nil
		},
	})
}

func TestAccClusterConnectorV2_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())
	p := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"tg": p},
		Steps: []resource.TestStep{
			{
				Config: clusterConnectorV2Config(testClusterFQDN, 100),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_cluster_connector.test", "connector_id"),
					resource.TestCheckResourceAttr("tg_cluster_connector.test", "cluster_fqdn", testClusterFQDN),
					resource.TestCheckResourceAttr("tg_cluster_connector.test", "port", "9090"),
					resource.TestCheckResourceAttr("tg_cluster_connector.test", "protocol", "tcp"),
					resource.TestCheckResourceAttr("tg_cluster_connector.test", "rate_limit", "100"),
					checkClusterConnectorAPISide(p, testClusterFQDN, 100),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_cluster_connector.test", tfjsonpath.New("connector_id")),
				},
			},
			{
				Config: clusterConnectorV2Config(testClusterFQDN, 500),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_cluster_connector.test", "rate_limit", "500"),
					checkClusterConnectorAPISide(p, testClusterFQDN, 500),
				),
			},
		},
	})
}

func clusterConnectorV2Config(clusterFQDN string, rateLimit int) string {
	return fmt.Sprintf(`
resource "tg_cluster_connector" "test" {
  cluster_fqdn = %q
  node         = "local"
  service      = "127.0.0.1:9091"
  port         = 9090
  protocol     = "tcp"
  description  = %q
  enabled      = true
  nic          = "any"
  rate_limit   = %d
}
`, clusterFQDN, v2ClusterTestConnectorDesc, rateLimit)
}

func checkClusterConnectorAPISide(p *schema.Provider, clusterFQDN string, wantRateLimit int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := p.Meta().(*tg.Client)
		rs, ok := s.RootModule().Resources["tg_cluster_connector.test"]
		if !ok {
			return fmt.Errorf("tg_cluster_connector.test not found in state")
		}
		connectorID := rs.Primary.ID

		var cluster tg.Cluster
		if err := client.Get(context.Background(), fmt.Sprintf("/cluster/%s", clusterFQDN), &cluster); err != nil {
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
