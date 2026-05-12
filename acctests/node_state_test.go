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

func TestAccNodeState_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	p := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": p,
		},
		Steps: []resource.TestStep{
			{
				Config: enabledNodeState(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_node_state.test", "id", testNodeID),
					resource.TestCheckResourceAttr("tg_node_state.test", "enabled", "true"),
					checkNodeStateAPISide(p, "ACTIVE"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_node_state.test", tfjsonpath.New("id")),
				},
			},
			{
				Config: disabledNodeState(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_node_state.test", "id", testNodeID),
					resource.TestCheckResourceAttr("tg_node_state.test", "enabled", "false"),
					checkNodeStateAPISide(p, "INACTIVE"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_node_state.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func enabledNodeState() string {
	return fmt.Sprintf(`
resource "tg_node_state" "test" {
  node_id = %q
  enabled = true
}
	`, testNodeID)
}

func disabledNodeState() string {
	return fmt.Sprintf(`
resource "tg_node_state" "test" {
  node_id = %q
  enabled = false
}
	`, testNodeID)
}

func checkNodeStateAPISide(p *schema.Provider, expected string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := p.Meta().(*tg.Client)
		var node tg.Node
		if err := client.Get(context.Background(), "/node/"+testNodeID, &node); err != nil {
			return fmt.Errorf("error fetching node: %w", err)
		}
		if node.State != expected {
			return fmt.Errorf("expected node state %q, got %q", expected, node.State)
		}
		return nil
	}
}
