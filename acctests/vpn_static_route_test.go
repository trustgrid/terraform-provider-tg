package acctests

import (
	"context"
	_ "embed"
	"fmt"
	"os"
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

//go:embed test-data/vpn-static-route/create.hcl
var vpnStaticRouteCreate string

//go:embed test-data/vpn-static-route/update.hcl
var vpnStaticRouteUpdate string

//go:embed test-data/vpn-static-route/create-cluster.hcl
var vpnClusterStaticRouteCreate string

//go:embed test-data/vpn-static-route/update-cluster.hcl
var vpnClusterStaticRouteUpdate string

func init() {
	resource.AddTestSweepers("tg_vpn_static_route", &resource.Sweeper{
		Name: "tg_vpn_static_route",
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

			routes := make([]tg.VPNRoute, 0)
			if err := client.Get(context.Background(), fmt.Sprintf("/v2/node/%s/vpn/%s/route", testNodeID, "test-vnet"), &routes); err != nil {
				return fmt.Errorf("error getting VPN static routes: %w", err)
			}

			// Just create this and wait for the other sweeper to clean up after you
			client.Post(context.Background(), "/v2/domain/"+client.Domain+"/network", tg.VirtualNetwork{Name: "test-vnet"})

			for _, r := range routes {
				if err := client.Delete(context.Background(), fmt.Sprintf("/v2/node/%s/vpn/%s/route/%s", testNodeID, "test-vnet", r.UID), nil); err != nil {
					return err
				}
			}

			if err := client.Get(context.Background(), fmt.Sprintf("/v2/cluster/%s/vpn/%s/route", testClusterFQDN, "test-vnet"), &routes); err != nil {
				return fmt.Errorf("error getting VPN static routes: %w", err)
			}

			for _, r := range routes {
				if err := client.Delete(context.Background(), fmt.Sprintf("/v2/cluster/%s/vpn/%s/route/%s", testClusterFQDN, "test-vnet", r.UID), nil); err != nil {
					return err
				}
			}

			return nil
		},
	})
}

func TestAccVPNStaticRoute_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: vpnStaticRouteCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_vpn_static_route.test", "id"),
					resource.TestCheckResourceAttr("tg_vpn_static_route.test", "network_name", "test-vnet"),
					resource.TestCheckResourceAttr("tg_vpn_static_route.test", "node_id", testNodeID),
					resource.TestCheckResourceAttr("tg_vpn_static_route.test", "node", "test-subject"),
					resource.TestCheckResourceAttr("tg_vpn_static_route.test", "network_cidr", "10.10.24.24/32"),
					resource.TestCheckResourceAttr("tg_vpn_static_route.test", "description", "Test VPN Static Route"),
					resource.TestCheckResourceAttr("tg_vpn_static_route.test", "path", "1.1.1.1"),
					resource.TestCheckResourceAttr("tg_vpn_static_route.test", "metric", "10"),
					checkVPNStaticRouteAPISide(provider, "test-subject", "Test VPN Static Route", "10.10.24.24/32", "1.1.1.1", 10),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_vpn_static_route.test", tfjsonpath.New("id")),
				},
			},
			{
				Config: vpnStaticRouteUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_vpn_static_route.test", "id"),
					resource.TestCheckResourceAttr("tg_vpn_static_route.test", "network_name", "test-vnet"),
					resource.TestCheckResourceAttr("tg_vpn_static_route.test", "node_id", testNodeID),
					resource.TestCheckResourceAttr("tg_vpn_static_route.test", "node", "another-subject"),
					resource.TestCheckResourceAttr("tg_vpn_static_route.test", "network_cidr", "10.10.24.0/24"),
					resource.TestCheckResourceAttr("tg_vpn_static_route.test", "description", "better description"),
					resource.TestCheckResourceAttr("tg_vpn_static_route.test", "path", "1.1.1.2"),
					resource.TestCheckResourceAttr("tg_vpn_static_route.test", "metric", "11"),
					checkVPNStaticRouteAPISide(provider, "another-subject", "better description", "10.10.24.0/24", "1.1.1.2", 11),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_vpn_static_route.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func TestAccVPNStaticRoute_ClusterHappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: vpnClusterStaticRouteCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_vpn_static_route.test", "id"),
					resource.TestCheckResourceAttr("tg_vpn_static_route.test", "network_name", "test-vnet"),
					resource.TestCheckResourceAttr("tg_vpn_static_route.test", "cluster_fqdn", testClusterFQDN),
					resource.TestCheckResourceAttr("tg_vpn_static_route.test", "node", "test-subject"),
					resource.TestCheckResourceAttr("tg_vpn_static_route.test", "network_cidr", "10.10.24.24/32"),
					resource.TestCheckResourceAttr("tg_vpn_static_route.test", "description", "Test VPN Static Route"),
					resource.TestCheckResourceAttr("tg_vpn_static_route.test", "metric", "10"),
					checkClusterVPNStaticRouteAPISide(provider, "test-subject", "Test VPN Static Route", "10.10.24.24/32", 10),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_vpn_static_route.test", tfjsonpath.New("id")),
				},
			},
			{
				Config: vpnClusterStaticRouteUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_vpn_static_route.test", "id"),
					resource.TestCheckResourceAttr("tg_vpn_static_route.test", "network_name", "test-vnet"),
					resource.TestCheckResourceAttr("tg_vpn_static_route.test", "cluster_fqdn", testClusterFQDN),
					resource.TestCheckResourceAttr("tg_vpn_static_route.test", "node", "another-subject"),
					resource.TestCheckResourceAttr("tg_vpn_static_route.test", "network_cidr", "10.10.24.0/24"),
					resource.TestCheckResourceAttr("tg_vpn_static_route.test", "description", "better description"),
					resource.TestCheckResourceAttr("tg_vpn_static_route.test", "metric", "11"),
					checkClusterVPNStaticRouteAPISide(provider, "another-subject", "better description", "10.10.24.0/24", 11),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_vpn_static_route.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func checkClusterVPNStaticRouteAPISide(provider *schema.Provider, node string, desc string, cidr string, metric int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)

		routes := make([]tg.VPNRoute, 0)
		if err := client.Get(context.Background(), fmt.Sprintf("/v2/cluster/%s/vpn/%s/route", testClusterFQDN, "test-vnet"), &routes); err != nil {
			return fmt.Errorf("error getting VPN static routes: %w", err)
		}
		var route tg.VPNRoute

		found := false
		for _, r := range routes {
			if r.Node == node && r.NetworkCIDR == cidr {
				found = true
				route = r
			}
		}

		if !found {
			return fmt.Errorf("static route not found")
		}

		if route.Description != desc {
			return fmt.Errorf("description mismatch: expected %s, got %s", desc, route.Description)
		}
		if route.NetworkCIDR != cidr {
			return fmt.Errorf("network CIDR mismatch: expected %s, got %s", cidr, route.NetworkCIDR)
		}
		if route.Metric != metric {
			return fmt.Errorf("metric mismatch: expected %d, got %d", metric, route.Metric)
		}

		return nil
	}
}

func checkVPNStaticRouteAPISide(provider *schema.Provider, node string, desc string, cidr string, path string, metric int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)

		routes := make([]tg.VPNRoute, 0)
		if err := client.Get(context.Background(), fmt.Sprintf("/v2/node/%s/vpn/%s/route", testNodeID, "test-vnet"), &routes); err != nil {
			return fmt.Errorf("error getting VPN static routes: %w", err)
		}
		var route tg.VPNRoute

		found := false
		for _, r := range routes {
			if r.Node == node && r.NetworkCIDR == cidr {
				found = true
				route = r
			}
		}

		if !found {
			return fmt.Errorf("static route not found")
		}

		if route.Description != desc {
			return fmt.Errorf("description mismatch: expected %s, got %s", desc, route.Description)
		}
		if route.NetworkCIDR != cidr {
			return fmt.Errorf("network CIDR mismatch: expected %s, got %s", cidr, route.NetworkCIDR)
		}
		if route.Path != path {
			return fmt.Errorf("path mismatch: expected %s, got %s", path, route.Path)
		}
		if route.Metric != metric {
			return fmt.Errorf("metric mismatch: expected %d, got %d", metric, route.Metric)
		}

		return nil
	}
}
