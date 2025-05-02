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

//go:embed test-data/vnet-group-membership/create.hcl
var vnetMembership string

func TestAccVirtualNetworkGroupMembership_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: vnetMembership,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_virtual_network_group_membership.test", "id"),
					resource.TestCheckResourceAttr("tg_virtual_network_group_membership.test", "object", "test-obj"),
					resource.TestCheckResourceAttr("tg_virtual_network_group_membership.test", "group", "test-group"),
					resource.TestCheckResourceAttr("tg_virtual_network_group_membership.test", "network", "test-membership"),
					checkVNetMembershipAPI(provider),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_virtual_network_group_membership.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func checkVNetMembershipAPI(provider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)

		var membership []tg.VNetGroupMembership
		if err := client.Get(context.Background(), "/v2/domain/"+client.Domain+"/network/test-membership/network-group/test-group", &membership); err != nil {
			return fmt.Errorf("error getting vnet group membership: %w", err)
		}

		for _, obj := range membership {
			if obj.Object == "test-obj" && obj.Group == "test-group" {
				return nil
			}
		}
		return fmt.Errorf("vnet membership not found")
	}
}
