package acctests

import (
	"context"
	"fmt"
	"os"
	"strings"
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

func init() {
	resource.AddTestSweepers("tg_virtual_network_route", &resource.Sweeper{
		Name:         "tg_virtual_network_route",
		Dependencies: []string{"tg_vpn_attachment"},
		F: func(r string) error {
			cp := tg.ClientParams{
				APIKey:    os.Getenv("TG_API_KEY_ID"),
				APISecret: os.Getenv("TG_API_KEY_SECRET"),
				APIHost:   os.Getenv("TG_API_HOST"),
			}
			client, err := tg.NewClient(context.Background(), cp)
			if err != nil {
				return fmt.Errorf("error creating client: %w", err)
			}

			var vnets []tg.VirtualNetwork
			if err := client.Get(context.Background(), "/v2/domain/"+client.Domain+"/network", &vnets); err != nil {
				return fmt.Errorf("error listing virtual networks: %w", err)
			}

			for _, vnet := range vnets {
				if !strings.HasPrefix(vnet.Name, testVNetPrefix) {
					continue
				}

				var routes []tg.VNetRoute
				if err := client.Get(context.Background(), "/v2/domain/"+client.Domain+"/network/"+vnet.Name+"/route", &routes); err != nil {
					return fmt.Errorf("error listing routes for %s: %w", vnet.Name, err)
				}

				for _, route := range routes {
					if err := client.Delete(context.Background(), "/v2/domain/"+client.Domain+"/network/"+vnet.Name+"/route/"+route.UID, nil); err != nil {
						return fmt.Errorf("error deleting route %s for %s: %w", route.UID, vnet.Name, err)
					}
				}
			}

			return nil
		},
	})
}

func TestAccVNetRoute_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())
	networkName := newTestVNetName("route-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	tgProvider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": tgProvider,
		},
		Steps: []resource.TestStep{
			{
				Config: vnetRouteConfig(networkName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_virtual_network_route.test", "id"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "network", networkName),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "dest", "test-cluster"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "network_cidr", "10.10.24.24/32"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "metric", "10"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "description", "Test Virtual Network Route"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "monitor.0.name", "tcp-probe"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "monitor.0.enabled", "true"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "monitor.0.protocol", "tcp"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "monitor.0.dest", "10.100.0.10"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "monitor.0.port", "443"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "monitor.0.interval", "5"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "monitor.0.count", "3"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "monitor.0.max_latency", "500"),
					checkVNetRouteAPISide(tgProvider, networkName, func(route tg.VNetRoute) error {
						if route.Dest != "test-cluster" {
							return fmt.Errorf("expected route dest test-cluster, got %s", route.Dest)
						}
						if len(route.Monitors) != 1 {
							return fmt.Errorf("expected 1 monitor, got %d", len(route.Monitors))
						}

						monitor := route.Monitors[0]
						if !monitor.Enabled || monitor.Name != "tcp-probe" || monitor.Protocol != "tcp" || monitor.Dest != "10.100.0.10" || monitor.Port == nil || *monitor.Port != 443 || monitor.Interval != 5 || monitor.Count != 3 || monitor.MaxLatency == nil || *monitor.MaxLatency != 500 {
							return fmt.Errorf("unexpected monitor: %+v", monitor)
						}

						return nil
					}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_virtual_network_route.test", tfjsonpath.New("id")),
				},
			},
			{
				Config: vnetRouteUpdatedConfig(networkName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_virtual_network_route.test", "id"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "dest", "test-cluster"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "network_cidr", "10.10.24.0/24"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "metric", "11"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "description", "Updated Test Virtual Network Route"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "monitor.0.name", "icmp-probe"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "monitor.0.enabled", "true"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "monitor.0.protocol", "icmp"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "monitor.0.dest", "10.100.0.11"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "monitor.0.interval", "10"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "monitor.0.count", "2"),
					resource.TestCheckNoResourceAttr("tg_virtual_network_route.test", "monitor.0.port"),
					checkVNetRouteAPISide(tgProvider, networkName, func(route tg.VNetRoute) error {
						if route.Dest != "test-cluster" {
							return fmt.Errorf("expected route dest test-cluster, got %s", route.Dest)
						}
						if len(route.Monitors) != 1 {
							return fmt.Errorf("expected 1 monitor, got %d", len(route.Monitors))
						}

						monitor := route.Monitors[0]
						if !monitor.Enabled || monitor.Name != "icmp-probe" || monitor.Protocol != "icmp" || monitor.Dest != "10.100.0.11" || monitor.Port != nil || monitor.Interval != 10 || monitor.Count != 2 {
							return fmt.Errorf("unexpected monitor: %+v", monitor)
						}

						return nil
					}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_virtual_network_route.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func vnetRouteConfig(networkName string) string {
	return fmt.Sprintf(`
resource "tg_virtual_network" "test" {
  name         = %q
  network_cidr = "10.10.0.0/16"
  description  = "Test Virtual Network"
  no_nat       = true
}

resource "tg_vpn_attachment" "test" {
  cluster_fqdn = %q
  network      = tg_virtual_network.test.name
}

resource "tg_virtual_network_route" "test" {
  network      = tg_vpn_attachment.test.network
  dest         = "test-cluster"
  network_cidr = "10.10.24.24/32"
  metric       = 10
  description  = "Test Virtual Network Route"

  monitor {
    name        = "tcp-probe"
    enabled     = true
    protocol    = "tcp"
    dest        = "10.100.0.10"
    port        = 443
    interval    = 5
    count       = 3
    max_latency = 500
  }
}
`, networkName, testClusterFQDN)
}

func vnetRouteUpdatedConfig(networkName string) string {
	return fmt.Sprintf(`
resource "tg_virtual_network" "test" {
  name         = %q
  network_cidr = "10.10.0.0/16"
  description  = "Test Virtual Network"
  no_nat       = true
}

resource "tg_vpn_attachment" "test" {
  cluster_fqdn = %q
  network      = tg_virtual_network.test.name
}

resource "tg_virtual_network_route" "test" {
  network      = tg_vpn_attachment.test.network
  dest         = "test-cluster"
  network_cidr = "10.10.24.0/24"
  metric       = 11
  description  = "Updated Test Virtual Network Route"

  monitor {
    name     = "icmp-probe"
    enabled  = true
    protocol = "icmp"
    dest     = "10.100.0.11"
    interval = 10
    count    = 2
  }
}
`, networkName, testClusterFQDN)
}

func checkVNetRouteAPISide(provider *schema.Provider, networkName string, check func(tg.VNetRoute) error) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)
		resourceState, ok := s.RootModule().Resources["tg_virtual_network_route.test"]
		if !ok {
			return fmt.Errorf("route resource not found in state")
		}

		routeID := resourceState.Primary.ID
		routes := make([]tg.VNetRoute, 0)
		if err := client.Get(context.Background(), "/v2/domain/"+client.Domain+"/network/"+networkName+"/route", &routes); err != nil {
			return fmt.Errorf("error getting virtual network routes: %w", err)
		}

		for _, route := range routes {
			if route.UID == routeID {
				return check(route)
			}
		}

		return fmt.Errorf("virtual network route %s not found", routeID)
	}
}
