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

const happyNetworkConfig = `
resource "tg_network_config" "test" {
  node_id = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
  dark_mode = true
  forwarding = true

  interface {
    nic = "ens192"
	dhcp = false
	gateway = "10.20.10.1"
	ip = "10.20.10.50/24"

	route {
	  route = "10.10.10.0/24"
	  description = "some desc"
	  next_hop = "127.0.0.1"
	}
  }
}`

func TestAccNetworkConfig_NodeHappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: happyNetworkConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_network_config.test", "id", "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"),
					resource.TestCheckResourceAttr("tg_network_config.test", "dark_mode", "true"),
					resource.TestCheckResourceAttr("tg_network_config.test", "forwarding", "true"),
					resource.TestCheckResourceAttr("tg_network_config.test", "interface.0.nic", "ens192"),
					resource.TestCheckResourceAttr("tg_network_config.test", "interface.0.dhcp", "false"),
					resource.TestCheckResourceAttr("tg_network_config.test", "interface.0.gateway", "10.20.10.1"),
					resource.TestCheckResourceAttr("tg_network_config.test", "interface.0.ip", "10.20.10.50/24"),
					resource.TestCheckResourceAttr("tg_network_config.test", "interface.0.route.0.route", "10.10.10.0/24"),
					resource.TestCheckResourceAttr("tg_network_config.test", "interface.0.route.0.description", "some desc"),
					resource.TestCheckResourceAttr("tg_network_config.test", "interface.0.route.0.next_hop", "127.0.0.1"),
					checkNetworkConfig(t.Context(), provider, "tg_network_config.test"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_network_config.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func checkNetworkConfig(ctx context.Context, provider *schema.Provider, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		client := provider.Meta().(*tg.Client)

		n := tg.Node{}
		if err := client.Get(ctx, "/node/"+rs.Primary.ID, &n); err != nil {
			return err
		}

		switch {
		case n.Config.Network.DarkMode == nil || *n.Config.Network.DarkMode != true:
			return fmt.Errorf("expected dark_mode to be true, got %v", n.Config.Network.DarkMode)
		case n.Config.Network.Forwarding == nil || *n.Config.Network.Forwarding != true:
			return fmt.Errorf("expected forwarding to be true, got %v", n.Config.Network.Forwarding)
		case len(n.Config.Network.Interfaces) != 1:
			return fmt.Errorf("expected 1 interfaces, got %d", len(n.Config.Network.Interfaces))
		case n.Config.Network.Interfaces[0].NIC != "ens192":
			return fmt.Errorf("expected NIC to be ens192, got %s", n.Config.Network.Interfaces[1].NIC)
		case n.Config.Network.Interfaces[0].DHCP != false:
			return fmt.Errorf("expected DHCP to be false, got %v", n.Config.Network.Interfaces[1].DHCP)
		case n.Config.Network.Interfaces[0].Gateway != "10.20.10.1":
			return fmt.Errorf("expected gateway to be 10.20.10.1, got %s", n.Config.Network.Interfaces[1].Gateway)
		case n.Config.Network.Interfaces[0].IP != "10.20.10.50/24":
			return fmt.Errorf("expected IP to be 10.20.10.50/24, got %s", n.Config.Network.Interfaces[1].IP)
		case len(n.Config.Network.Interfaces[0].Routes) != 1:
			return fmt.Errorf("expected 1 route, got %d", len(n.Config.Network.Interfaces[1].Routes))
		case n.Config.Network.Interfaces[0].Routes[0].Route != "10.10.10.0/24":
			return fmt.Errorf("expected route to be 10.10.10.0/24, got %s", n.Config.Network.Interfaces[1].Routes[0].Route)
		case n.Config.Network.Interfaces[0].Routes[0].Description != "some desc":
			return fmt.Errorf("expected route description to be 'some desc', got %s", n.Config.Network.Interfaces[1].Routes[0].Description)
		case n.Config.Network.Interfaces[0].Routes[0].Next != "127.0.0.1":
			return fmt.Errorf("expected route next to be '127.0.0.1', got %s", n.Config.Network.Interfaces[1].Routes[0].Next)
		}

		return nil
	}
}
