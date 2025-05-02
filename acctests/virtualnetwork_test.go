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

func TestAccVirtualNetwork_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: virtualNetworkConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_virtual_network.test", "id"),
					resource.TestCheckResourceAttr("tg_virtual_network.test", "name", "test-vnet"),
					resource.TestCheckResourceAttr("tg_virtual_network.test", "network_cidr", "10.10.0.0/16"),
					resource.TestCheckResourceAttr("tg_virtual_network.test", "description", "Test Virtual Network"),
					resource.TestCheckResourceAttr("tg_virtual_network.test", "no_nat", "true"),
					checkVirtualNetworkAPISide(provider, "test-vnet", true),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_virtual_network.test", tfjsonpath.New("id")),
				},
			},
			{
				Config: virtualNetworkUpdatedConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_virtual_network.test", "id"),
					resource.TestCheckResourceAttr("tg_virtual_network.test", "name", "test-vnet"),
					resource.TestCheckResourceAttr("tg_virtual_network.test", "network_cidr", "10.20.0.0/16"),
					resource.TestCheckResourceAttr("tg_virtual_network.test", "description", "Updated Test Virtual Network"),
					resource.TestCheckResourceAttr("tg_virtual_network.test", "no_nat", "false"),
					checkVirtualNetworkAPISide(provider, "test-vnet", false),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_virtual_network.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func virtualNetworkConfig() string {
	return `
resource "tg_virtual_network" "test" {
  name         = "test-vnet"
  network_cidr = "10.10.0.0/16"
  description  = "Test Virtual Network"
  no_nat       = true
}
	`
}

func virtualNetworkUpdatedConfig() string {
	return `
resource "tg_virtual_network" "test" {
  name         = "test-vnet"
  network_cidr = "10.20.0.0/16"
  description  = "Updated Test Virtual Network"
  no_nat       = false
}
	`
}

func checkVirtualNetworkAPISide(provider *schema.Provider, name string, nonat bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)

		vnets := make([]tg.VirtualNetwork, 0)
		if err := client.Get(context.Background(), "/v2/domain/"+client.Domain+"/network", &vnets); err != nil {
			return fmt.Errorf("error getting virtual networks: %w", err)
		}

		var vnet tg.VirtualNetwork
		found := false
		for _, v := range vnets {
			if v.Name == name {
				vnet = v
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("virtual network %s not found", name)
		}

		if vnet.Name != name {
			return fmt.Errorf("expected virtual network name %s, got %s", name, vnet.Name)
		}
		if vnet.NoNAT != nonat {
			return fmt.Errorf("expected nonat %t got %t", nonat, vnet.NoNAT)
		}

		return nil
	}
}
