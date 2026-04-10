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

func TestAccVirtualNetworkGroup_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())
	networkName := newTestVNetName("group-network-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))
	groupName := newTestVNetName("group-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: vnetGroupConfig(networkName, groupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_virtual_network_group.test", "id"),
					resource.TestCheckResourceAttr("tg_virtual_network_group.test", "name", groupName),
					checkVNetGroupAPI(provider, networkName, groupName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_virtual_network_group.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func vnetGroupConfig(networkName string, groupName string) string {
	return fmt.Sprintf(`
resource "tg_virtual_network" "group_network" {
  name        = "%s"
  description = "Group Test Virtual Network"
  no_nat      = false
}

resource "tg_virtual_network_group" "test" {
  name    = "%s"
  network = resource.tg_virtual_network.group_network.name
}
	`, networkName, groupName)
}

func checkVNetGroupAPI(provider *schema.Provider, networkName string, groupName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)

		var groups []tg.VNetGroup
		if err := client.Get(context.Background(), "/v2/domain/"+client.Domain+"/network/"+networkName+"/network-group", &groups); err != nil {
			return fmt.Errorf("error getting vnet groups: %w", err)
		}

		for _, obj := range groups {
			if obj.Name == groupName {
				return nil
			}
		}

		return fmt.Errorf("vnet group not found")
	}
}
