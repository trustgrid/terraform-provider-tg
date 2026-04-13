package acctests

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/trustgrid/terraform-provider-tg/provider"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

// ---------------------------------------------------------------------------
// tg_node_interface
// ---------------------------------------------------------------------------

func nodeInterfaceConfig(ip string) string {
	return fmt.Sprintf(`
resource "tg_node_interface" "test" {
  node_id = %q
  nic     = "ens192"
  ip      = %q
  gateway = "10.20.10.1"
  dhcp    = false
}
`, testNodeID, ip)
}

func TestAccNodeInterface_HappyPath(t *testing.T) {
	p := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"tg": p},
		Steps: []resource.TestStep{
			{
				Config: nodeInterfaceConfig("10.20.10.50/24"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_node_interface.test", "nic", "ens192"),
					resource.TestCheckResourceAttr("tg_node_interface.test", "ip", "10.20.10.50/24"),
					resource.TestCheckResourceAttr("tg_node_interface.test", "gateway", "10.20.10.1"),
					resource.TestCheckResourceAttr("tg_node_interface.test", "dhcp", "false"),
					checkNodeInterfaceInAPI(t.Context(), p, "tg_node_interface.test"),
				),
			},
			{
				// Update IP — verifies Update works.
				Config: nodeInterfaceConfig("10.20.10.51/24"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_node_interface.test", "ip", "10.20.10.51/24"),
					checkNodeInterfaceInAPI(t.Context(), p, "tg_node_interface.test"),
				),
			},
		},
	})
}

func checkNodeInterfaceInAPI(ctx context.Context, p *schema.Provider, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		nodeID := rs.Primary.Attributes["node_id"]
		nic := rs.Primary.Attributes["nic"]
		expectedIP := rs.Primary.Attributes["ip"]

		client := p.Meta().(*tg.Client)
		n := tg.Node{}
		if err := client.Get(ctx, "/node/"+nodeID, &n); err != nil {
			return err
		}

		for _, iface := range n.Config.Network.Interfaces {
			if iface.NIC == nic {
				if iface.IP != expectedIP {
					return fmt.Errorf("interface %s: expected IP %q, got %q", nic, expectedIP, iface.IP)
				}
				return nil
			}
		}
		return fmt.Errorf("interface %s not found in network config", nic)
	}
}

// ---------------------------------------------------------------------------
// tg_node_interface_route
// ---------------------------------------------------------------------------

func nodeInterfaceRouteConfig(nextHop string) string {
	return fmt.Sprintf(`
resource "tg_node_interface" "route_base" {
  node_id = %q
  nic     = "ens192"
  ip      = "10.20.10.50/24"
  gateway = "10.20.10.1"
  dhcp    = false
}

resource "tg_node_interface_route" "test" {
  node_id     = %q
  nic         = "ens192"
  route       = "10.10.10.0/24"
  next_hop    = %q
  description = "acceptance test route"
  depends_on  = [tg_node_interface.route_base]
}
`, testNodeID, testNodeID, nextHop)
}

func TestAccNodeInterfaceRoute_HappyPath(t *testing.T) {
	p := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"tg": p},
		Steps: []resource.TestStep{
			{
				Config: nodeInterfaceRouteConfig("10.20.10.1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_node_interface_route.test", "nic", "ens192"),
					resource.TestCheckResourceAttr("tg_node_interface_route.test", "route", "10.10.10.0/24"),
					resource.TestCheckResourceAttr("tg_node_interface_route.test", "next_hop", "10.20.10.1"),
					resource.TestCheckResourceAttr("tg_node_interface_route.test", "description", "acceptance test route"),
					checkNodeInterfaceRouteInAPI(t.Context(), p, "tg_node_interface_route.test"),
				),
			},
			{
				// Update next_hop — verifies Update works.
				Config: nodeInterfaceRouteConfig("10.20.10.2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_node_interface_route.test", "next_hop", "10.20.10.2"),
					checkNodeInterfaceRouteInAPI(t.Context(), p, "tg_node_interface_route.test"),
				),
			},
		},
	})
}

func checkNodeInterfaceRouteInAPI(ctx context.Context, p *schema.Provider, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		nodeID := rs.Primary.Attributes["node_id"]
		nic := rs.Primary.Attributes["nic"]
		dest := rs.Primary.Attributes["route"]
		expectedNextHop := rs.Primary.Attributes["next_hop"]

		client := p.Meta().(*tg.Client)
		n := tg.Node{}
		if err := client.Get(ctx, "/node/"+nodeID, &n); err != nil {
			return err
		}

		for _, iface := range n.Config.Network.Interfaces {
			if iface.NIC != nic {
				continue
			}
			for _, r := range iface.Routes {
				if r.Route == dest {
					if r.Next != expectedNextHop {
						return fmt.Errorf("route %s on %s: expected next_hop %q, got %q", dest, nic, expectedNextHop, r.Next)
					}
					return nil
				}
			}
			return fmt.Errorf("route %s not found on interface %s", dest, nic)
		}
		return fmt.Errorf("interface %s not found in network config", nic)
	}
}

// ---------------------------------------------------------------------------
// tg_node_interface_vlan
// ---------------------------------------------------------------------------

func nodeInterfaceVLANConfig(ip string) string {
	return fmt.Sprintf(`
resource "tg_node_interface" "vlan_base" {
  node_id = %q
  nic     = "ens192"
  ip      = "10.20.10.50/24"
  gateway = "10.20.10.1"
  dhcp    = false
}

resource "tg_node_interface_vlan" "test" {
  node_id     = %q
  nic         = "ens192"
  vlan_id     = 100
  ip          = %q
  description = "acceptance test vlan"
  depends_on  = [tg_node_interface.vlan_base]
}
`, testNodeID, testNodeID, ip)
}

func TestAccNodeInterfaceVLAN_HappyPath(t *testing.T) {
	p := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"tg": p},
		Steps: []resource.TestStep{
			{
				Config: nodeInterfaceVLANConfig("192.168.100.1/24"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_node_interface_vlan.test", "nic", "ens192"),
					resource.TestCheckResourceAttr("tg_node_interface_vlan.test", "vlan_id", "100"),
					resource.TestCheckResourceAttr("tg_node_interface_vlan.test", "ip", "192.168.100.1/24"),
					resource.TestCheckResourceAttr("tg_node_interface_vlan.test", "description", "acceptance test vlan"),
					checkNodeInterfaceVLANInAPI(t.Context(), p, "tg_node_interface_vlan.test"),
				),
			},
			{
				// Update IP.
				Config: nodeInterfaceVLANConfig("192.168.100.2/24"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_node_interface_vlan.test", "ip", "192.168.100.2/24"),
					checkNodeInterfaceVLANInAPI(t.Context(), p, "tg_node_interface_vlan.test"),
				),
			},
		},
	})
}

func checkNodeInterfaceVLANInAPI(ctx context.Context, p *schema.Provider, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		nodeID := rs.Primary.Attributes["node_id"]
		nic := rs.Primary.Attributes["nic"]
		expectedIP := rs.Primary.Attributes["ip"]
		var expectedVLANID int
		if _, err := fmt.Sscanf(rs.Primary.Attributes["vlan_id"], "%d", &expectedVLANID); err != nil {
			return fmt.Errorf("invalid vlan_id in state: %w", err)
		}

		client := p.Meta().(*tg.Client)
		n := tg.Node{}
		if err := client.Get(ctx, "/node/"+nodeID, &n); err != nil {
			return err
		}

		for _, iface := range n.Config.Network.Interfaces {
			if iface.NIC != nic {
				continue
			}
			for _, sub := range iface.SubInterfaces {
				if sub.VLANID == expectedVLANID {
					if sub.IP != expectedIP {
						return fmt.Errorf("vlan %d on %s: expected IP %q, got %q", expectedVLANID, nic, expectedIP, sub.IP)
					}
					return nil
				}
			}
			return fmt.Errorf("vlan %d not found on interface %s", expectedVLANID, nic)
		}
		return fmt.Errorf("interface %s not found in network config", nic)
	}
}
