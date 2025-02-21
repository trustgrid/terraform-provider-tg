package acctests

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/trustgrid/terraform-provider-tg/provider"
)

func TestAccNodeState_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider.New("test")(),
		},
		Steps: []resource.TestStep{
			{
				Config: enabledNodeState(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_node_state.test", "id", "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"),
					resource.TestCheckResourceAttr("tg_node_state.test", "enabled", "true"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_node_state.test", tfjsonpath.New("id")),
				},
			},
			{
				Config: disabledNodeState(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_node_state.test", "id", "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"),
					resource.TestCheckResourceAttr("tg_node_state.test", "enabled", "false"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_node_state.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func enabledNodeState() string {
	return `
resource "tg_node_state" "test" {
  node_id = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
  enabled = true
}
	`
}

func disabledNodeState() string {
	return `
resource "tg_node_state" "test" {
  node_id = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
  enabled = false
}
	`
}
