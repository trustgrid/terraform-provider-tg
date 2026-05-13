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

const v2NodeTestServiceName = "tf-test-v2-node-svc"

func init() {
	resource.AddTestSweepers("tg_node_service", &resource.Sweeper{
		Name: "tg_node_service",
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
				return fmt.Errorf("error fetching node %s: %w", testNodeID, err)
			}
			for _, svc := range node.Config.Services.Services {
				if svc.Name != v2NodeTestServiceName {
					continue
				}
				url := fmt.Sprintf("/v2/node/%s/config/services/%s", testNodeID, svc.ID)
				if err := client.Delete(context.Background(), url, nil); err != nil {
					return fmt.Errorf("error deleting node service %s: %w", svc.ID, err)
				}
			}
			return nil
		},
	})
}

func TestAccNodeServiceV2_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())
	p := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"tg": p},
		Steps: []resource.TestStep{
			{
				Config: nodeServiceV2Config(testNodeID, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_node_service.test", "service_id"),
					resource.TestCheckResourceAttr("tg_node_service.test", "node_id", testNodeID),
					resource.TestCheckResourceAttr("tg_node_service.test", "name", v2NodeTestServiceName),
					resource.TestCheckResourceAttr("tg_node_service.test", "protocol", "tcp"),
					resource.TestCheckResourceAttr("tg_node_service.test", "host", "10.0.0.2"),
					resource.TestCheckResourceAttr("tg_node_service.test", "port", "8081"),
					checkNodeServiceAPISide(p, testNodeID, v2NodeTestServiceName, ""),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_node_service.test", tfjsonpath.New("service_id")),
				},
			},
			{
				Config: nodeServiceV2Config(testNodeID, "ens192"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_node_service.test", "source_interface", "ens192"),
					checkNodeServiceAPISide(p, testNodeID, v2NodeTestServiceName, "ens192"),
				),
			},
		},
	})
}

func nodeServiceV2Config(nodeID, sourceInterface string) string {
	srcLine := ""
	if sourceInterface != "" {
		srcLine = fmt.Sprintf(`  source_interface = %q`, sourceInterface)
	}
	return fmt.Sprintf(`
resource "tg_node_service" "test" {
  node_id  = %q
  name     = %q
  protocol = "tcp"
  host     = "10.0.0.2"
  port     = 8081
  enabled  = true
%s
}
`, nodeID, v2NodeTestServiceName, srcLine)
}

func checkNodeServiceAPISide(p *schema.Provider, nodeID, wantName, wantSourceInterface string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := p.Meta().(*tg.Client)
		rs, ok := s.RootModule().Resources["tg_node_service.test"]
		if !ok {
			return fmt.Errorf("tg_node_service.test not found in state")
		}
		serviceID := rs.Primary.ID

		var node tg.Node
		if err := client.Get(context.Background(), fmt.Sprintf("/node/%s", nodeID), &node); err != nil {
			return fmt.Errorf("error getting node: %w", err)
		}
		var found *tg.Service
		for i := range node.Config.Services.Services {
			if node.Config.Services.Services[i].ID == serviceID {
				found = &node.Config.Services.Services[i]
				break
			}
		}
		if found == nil {
			return fmt.Errorf("service %s not found on node", serviceID)
		}
		if found.Name != wantName {
			return fmt.Errorf("expected name %q, got %q", wantName, found.Name)
		}
		if found.SourceInterface != wantSourceInterface {
			return fmt.Errorf("expected source_interface %q, got %q", wantSourceInterface, found.SourceInterface)
		}
		return nil
	}
}
