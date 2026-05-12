package acctests

import (
	"context"
	"fmt"
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

func TestAccConnector_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	p := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": p,
		},
		Steps: []resource.TestStep{
			{
				Config: connectorConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_connector.tomcat", "node_id", testNodeID),
					resource.TestCheckResourceAttr("tg_connector.tomcat", "node", "local"),
					resource.TestCheckResourceAttr("tg_connector.tomcat", "service", "127.0.0.1:8080"),
					resource.TestCheckResourceAttr("tg_connector.tomcat", "port", "8081"),
					resource.TestCheckResourceAttr("tg_connector.tomcat", "protocol", "tcp"),
					resource.TestCheckResourceAttr("tg_connector.tomcat", "description", "tomcat forwarding connector"),
					resource.TestCheckResourceAttr("tg_connector.tomcat", "nic", "eth0"),
					checkConnectorAPISide(p),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_connector.tomcat", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func connectorConfig() string {
	return fmt.Sprintf(`
resource "tg_connector" "tomcat" {
  node_id     = %q
  node        = "local"
  service     = "127.0.0.1:8080"
  port        = 8081
  protocol    = "tcp"
  description = "tomcat forwarding connector"
  nic = "eth0"
}`, testNodeID)
}

func checkConnectorAPISide(p *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := p.Meta().(*tg.Client)

		rs, ok := s.RootModule().Resources["tg_connector.tomcat"]
		if !ok {
			return fmt.Errorf("tg_connector.tomcat not found in state")
		}

		var node tg.Node
		if err := client.Get(context.Background(), "/node/"+testNodeID, &node); err != nil {
			return fmt.Errorf("error fetching node: %w", err)
		}

		var found *tg.Connector
		for i := range node.Config.Connectors.Connectors {
			c := &node.Config.Connectors.Connectors[i]
			if c.ID == rs.Primary.ID {
				found = c
				break
			}
		}
		if found == nil {
			return fmt.Errorf("connector %s not found on node (have %d)", rs.Primary.ID, len(node.Config.Connectors.Connectors))
		}

		if found.Node != "local" {
			return fmt.Errorf("expected node local, got %q", found.Node)
		}
		if found.Service != "127.0.0.1:8080" {
			return fmt.Errorf("expected service 127.0.0.1:8080, got %q", found.Service)
		}
		if found.Port != 8081 {
			return fmt.Errorf("expected port 8081, got %d", found.Port)
		}
		if found.Protocol != "tcp" {
			return fmt.Errorf("expected protocol tcp, got %q", found.Protocol)
		}
		if found.Description != "tomcat forwarding connector" {
			return fmt.Errorf("expected description, got %q", found.Description)
		}
		if found.NIC != "eth0" {
			return fmt.Errorf("expected nic eth0, got %q", found.NIC)
		}
		return nil
	}
}
