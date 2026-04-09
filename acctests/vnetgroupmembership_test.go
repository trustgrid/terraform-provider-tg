package acctests

import (
	"context"
	"fmt"
	"testing"

	_ "embed"

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

func TestAccVirtualNetworkGroupMembership_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	provider := provider.New("test")()
	networkName := "tf-test-membership-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	groupName := "tf-test-group-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	objectName := "tf-test-obj-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: vnetMembershipConfig(networkName, groupName, objectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_virtual_network_group_membership.test", "id"),
					resource.TestCheckResourceAttr("tg_virtual_network_group_membership.test", "object", objectName),
					resource.TestCheckResourceAttr("tg_virtual_network_group_membership.test", "group", groupName),
					resource.TestCheckResourceAttr("tg_virtual_network_group_membership.test", "network", networkName),
					checkVNetMembershipAPI(provider, networkName, groupName, objectName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_virtual_network_group_membership.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func vnetMembershipConfig(networkName string, groupName string, objectName string) string {
	return fmt.Sprintf(`
resource "tg_virtual_network" "member_test" {
  name = %q
}

resource "tg_virtual_network_group" "test" {
  name    = %q
  network = tg_virtual_network.member_test.name
}

resource "tg_virtual_network_object" "test" {
  name    = %q
  cidr    = "10.10.20.0/24"
  network = tg_virtual_network.member_test.name
}

resource "tg_virtual_network_group_membership" "test" {
  object  = tg_virtual_network_object.test.name
  group   = tg_virtual_network_group.test.name
  network = tg_virtual_network.member_test.name
}
`, networkName, groupName, objectName)
}

func checkVNetMembershipAPI(provider *schema.Provider, networkName string, groupName string, objectName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)

		var membership []tg.VNetGroupMembership
		if err := client.Get(context.Background(), "/v2/domain/"+client.Domain+"/network/"+networkName+"/network-group/"+groupName, &membership); err != nil {
			return fmt.Errorf("error getting vnet group membership: %w", err)
		}

		for _, obj := range membership {
			if obj.Object == objectName && obj.Group == groupName {
				return nil
			}
		}
		return fmt.Errorf("vnet membership not found")
	}
}
