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

const v2NodeTestConnectorDesc = "tf-test-v2-node-conn"

func init() {
	resource.AddTestSweepers("tg_node_connector", &resource.Sweeper{
		Name: "tg_node_connector",
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

			var node tg.Node
			if err := client.Get(context.Background(), "/node/"+testNodeID, &node); err != nil {
				return fmt.Errorf("error fetching node: %w", err)
			}
			for _, conn := range node.Config.Connectors.Connectors {
				if conn.Description != v2NodeTestConnectorDesc {
					continue
				}
				url := fmt.Sprintf("/v2/node/%s/config/connectors/%s", testNodeID, conn.ID)
				if err := client.Delete(context.Background(), url, nil); err != nil {
					return fmt.Errorf("error deleting node connector %s: %w", conn.ID, err)
				}
			}
			return nil
		},
	})
}

func TestAccNodeConnectorV2_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())
	p := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"tg": p},
		Steps: []resource.TestStep{
			{
				Config: nodeConnectorV2Config(testNodeID, 0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_node_connector.test", "connector_id"),
					resource.TestCheckResourceAttr("tg_node_connector.test", "node_id", testNodeID),
					resource.TestCheckResourceAttr("tg_node_connector.test", "port", "9092"),
					resource.TestCheckResourceAttr("tg_node_connector.test", "protocol", "tcp"),
					checkNodeConnectorAPISide(p, testNodeID, 0),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_node_connector.test", tfjsonpath.New("connector_id")),
				},
			},
			{
				Config: nodeConnectorV2Config(testNodeID, 750),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_node_connector.test", "rate_limit", "750"),
					checkNodeConnectorAPISide(p, testNodeID, 750),
				),
			},
		},
	})
}

func nodeConnectorV2Config(nodeID string, rateLimit int) string {
	return fmt.Sprintf(`
# Ensure the node has been upgraded to V2 connectors config before creating
# any V2 connectors. Idempotent — no-op if already V2.
resource "tg_node_connectors_v2_upgrade" "test" {
  node_id = %q
}

resource "tg_node_connector" "test" {
  node_id      = %q
  depends_on   = [tg_node_connectors_v2_upgrade.test]
  node         = "local"
  service      = "127.0.0.1:9093"
  port         = 9092
  protocol     = "tcp"
  description  = %q
  enabled      = true
  nic          = "any"
  rate_limit   = %d
}
`, nodeID, nodeID, v2NodeTestConnectorDesc, rateLimit)
}

func checkNodeConnectorAPISide(p *schema.Provider, nodeID string, wantRateLimit int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := p.Meta().(*tg.Client)
		rs, ok := s.RootModule().Resources["tg_node_connector.test"]
		if !ok {
			return fmt.Errorf("tg_node_connector.test not found in state")
		}
		connectorID := rs.Primary.ID

		var node tg.Node
		if err := client.Get(context.Background(), fmt.Sprintf("/node/%s", nodeID), &node); err != nil {
			return fmt.Errorf("error getting node: %w", err)
		}
		var found *tg.Connector
		for i := range node.Config.Connectors.Connectors {
			if node.Config.Connectors.Connectors[i].ID == connectorID {
				found = &node.Config.Connectors.Connectors[i]
				break
			}
		}
		if found == nil {
			return fmt.Errorf("connector %s not found on node", connectorID)
		}
		if found.RateLimit != wantRateLimit {
			return fmt.Errorf("expected rate_limit %d, got %d", wantRateLimit, found.RateLimit)
		}
		return nil
	}
}
