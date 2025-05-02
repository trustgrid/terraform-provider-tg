package acctests

import (
	"context"
	"fmt"
	"testing"

	_ "embed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/trustgrid/terraform-provider-tg/provider"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

//go:embed test-data/vnet-group/create.hcl
var vnetGroupCreate string

func TestAccVirtualNetworkGroup_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: vnetGroupCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_virtual_network_group.test", "id"),
					resource.TestCheckResourceAttr("tg_virtual_network_group.test", "name", "test-group"),
					checkVNetGroupAPI(provider),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_virtual_network_group.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func checkVNetGroupAPI(provider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)

		var groups []tg.VNetGroup
		if err := client.Get(context.Background(), "/v2/domain/"+client.Domain+"/network/test-group/network-group", &groups); err != nil {
			return fmt.Errorf("error getting vnet groups: %w", err)
		}

		for _, obj := range groups {
			if obj.Name == "test-group" {
				return nil
			}
		}

		return fmt.Errorf("vnet group not found")
	}
}
