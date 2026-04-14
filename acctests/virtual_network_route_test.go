package acctests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/trustgrid/terraform-provider-tg/provider"
)

func TestAccVirtualNetworkRoute_HappyPath(t *testing.T) {
	networkName := newTestVNetName("route-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	providerFactory := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": providerFactory,
		},
		Steps: []resource.TestStep{
			{
				Config: virtualNetworkRouteConfig(networkName, "basic"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_virtual_network_route.test", "uid"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "network", networkName),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "dest", "edge-node"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "network_cidr", "10.100.0.0/24"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "metric", "100"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "description", "basic route"),
				),
			},
		},
	})
}

func TestAccVirtualNetworkRoute_WithMonitor(t *testing.T) {
	networkName := newTestVNetName("route-mon-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	providerFactory := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": providerFactory,
		},
		Steps: []resource.TestStep{
			{
				Config: virtualNetworkRouteWithMonitorConfig(networkName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_virtual_network_route.test", "uid"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "network", networkName),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "monitor.#", "1"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "monitor.0.name", "tcp-probe"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "monitor.0.enabled", "true"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "monitor.0.protocol", "tcp"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "monitor.0.dest", "10.100.0.10"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "monitor.0.port", "443"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "monitor.0.interval", "5"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "monitor.0.count", "3"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "monitor.0.max_latency", "500"),
				),
			},
			{
				Config: virtualNetworkRouteWithMonitorUpdatedConfig(networkName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_virtual_network_route.test", "uid"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "monitor.#", "1"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "monitor.0.interval", "10"),
					resource.TestCheckResourceAttr("tg_virtual_network_route.test", "monitor.0.count", "5"),
				),
			},
		},
	})
}

func virtualNetworkRouteConfig(networkName, description string) string {
	return fmt.Sprintf(`
resource "tg_virtual_network" "test" {
  name        = "%s"
  description = "Test Virtual Network for route"
  no_nat      = true
}

resource "tg_virtual_network_route" "test" {
  network      = tg_virtual_network.test.name
  dest         = "edge-node"
  network_cidr = "10.100.0.0/24"
  metric       = 100
  description  = "%s"
}
	`, networkName, description)
}

func virtualNetworkRouteWithMonitorConfig(networkName string) string {
	return fmt.Sprintf(`
resource "tg_virtual_network" "test" {
  name        = "%s"
  description = "Test Virtual Network for route with monitor"
  no_nat      = true
}

resource "tg_virtual_network_route" "test" {
  network      = tg_virtual_network.test.name
  dest         = "edge-node"
  network_cidr = "10.100.0.0/24"
  metric       = 100
  description  = "route with tcp monitor"

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
	`, networkName)
}

func virtualNetworkRouteWithMonitorUpdatedConfig(networkName string) string {
	return fmt.Sprintf(`
resource "tg_virtual_network" "test" {
  name        = "%s"
  description = "Test Virtual Network for route with monitor"
  no_nat      = true
}

resource "tg_virtual_network_route" "test" {
  network      = tg_virtual_network.test.name
  dest         = "edge-node"
  network_cidr = "10.100.0.0/24"
  metric       = 100
  description  = "route with tcp monitor updated"

  monitor {
    name        = "tcp-probe"
    enabled     = true
    protocol    = "tcp"
    dest        = "10.100.0.10"
    port        = 443
    interval    = 10
    count       = 5
    max_latency = 500
  }
}
	`, networkName)
}
