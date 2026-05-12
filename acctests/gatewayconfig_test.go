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

func TestAccGatewayConfig_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	p := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": p,
		},
		Steps: []resource.TestStep{
			{
				Config: gatewayConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_gateway_config.test", "id", testNodeID),
					resource.TestCheckResourceAttr("tg_gateway_config.test", "enabled", "true"),
					resource.TestCheckResourceAttr("tg_gateway_config.test", "node_id", testNodeID),
					resource.TestCheckResourceAttr("tg_gateway_config.test", "host", "10.10.10.10"),
					resource.TestCheckResourceAttr("tg_gateway_config.test", "port", "8553"),
					resource.TestCheckResourceAttr("tg_gateway_config.test", "type", "public"),
					resource.TestCheckResourceAttr("tg_gateway_config.test", "udp_enabled", "true"),
					resource.TestCheckResourceAttr("tg_gateway_config.test", "udp_port", "5555"),
					resource.TestCheckResourceAttr("tg_gateway_config.test", "maxmbps", "50"),
					resource.TestCheckResourceAttr("tg_gateway_config.test", "max_client_write_mbps", "10"),
					resource.TestCheckResourceAttr("tg_gateway_config.test", "monitor_hops", "true"),
					resource.TestCheckResourceAttr("tg_gateway_config.test", "path.0.id", "anodepath"),
					resource.TestCheckResourceAttr("tg_gateway_config.test", "path.0.host", "5.5.5.5"),
					resource.TestCheckResourceAttr("tg_gateway_config.test", "path.0.port", "1234"),
					resource.TestCheckResourceAttr("tg_gateway_config.test", "path.0.node", "anode"),
					resource.TestCheckResourceAttr("tg_gateway_config.test", "path.0.default", "false"),
					resource.TestCheckResourceAttr("tg_gateway_config.test", "path.0.enabled", "true"),
					resource.TestCheckResourceAttr("tg_gateway_config.test", "path.0.local", "6.6.6.6"),
					resource.TestCheckResourceAttr("tg_gateway_config.test", "route.0.route", "test-subject"),
					resource.TestCheckResourceAttr("tg_gateway_config.test", "route.0.dest", "somewhere"),
					resource.TestCheckResourceAttr("tg_gateway_config.test", "route.0.metric", "3"),
					checkGatewayConfigAPISide(p),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_gateway_config.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func gatewayConfig() string {
	return fmt.Sprintf(`
resource "tg_gateway_config" "test" {
  node_id = %q
  enabled = true
  host = "10.10.10.10"
  port = 8553
  type = "public"
  udp_enabled = true
  udp_port = 5555
  maxmbps = 50
  monitor_hops = true
  max_client_write_mbps = 10
  path {
	id = "anodepath"
    host = "5.5.5.5"
	port = 1234
	node = "anode"
	local = "6.6.6.6"
  }

  route {
	route = "test-subject"
    dest = "somewhere"
	metric = 3
  }
}
	`, testNodeID)
}

func checkGatewayConfigAPISide(p *schema.Provider) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := p.Meta().(*tg.Client)

		var node tg.Node
		if err := client.Get(context.Background(), "/node/"+testNodeID, &node); err != nil {
			return fmt.Errorf("error fetching node: %w", err)
		}
		gw := node.Config.Gateway

		if !gw.Enabled {
			return fmt.Errorf("expected gateway enabled")
		}
		if gw.Host != "10.10.10.10" {
			return fmt.Errorf("expected host 10.10.10.10, got %q", gw.Host)
		}
		if gw.Port != 8553 {
			return fmt.Errorf("expected port 8553, got %d", gw.Port)
		}
		if gw.Type != "public" {
			return fmt.Errorf("expected type public, got %q", gw.Type)
		}
		if !gw.UDPEnabled {
			return fmt.Errorf("expected udp_enabled")
		}
		if gw.UDPPort != 5555 {
			return fmt.Errorf("expected udp_port 5555, got %d", gw.UDPPort)
		}
		if gw.MaxMBPS != 50 {
			return fmt.Errorf("expected maxmbps 50, got %d", gw.MaxMBPS)
		}
		if gw.MaxClientWriteMBPS != 10 {
			return fmt.Errorf("expected max_client_write_mbps 10, got %d", gw.MaxClientWriteMBPS)
		}
		if !gw.MonitorHops {
			return fmt.Errorf("expected monitor_hops")
		}
		if len(gw.Paths) != 1 {
			return fmt.Errorf("expected 1 path, got %d", len(gw.Paths))
		}
		path := gw.Paths[0]
		if path.ID != "anodepath" || path.Host != "5.5.5.5" || path.Port != 1234 || path.Node != "anode" || path.Local != "6.6.6.6" {
			return fmt.Errorf("path mismatch: %+v", path)
		}
		if len(gw.Routes) != 1 {
			return fmt.Errorf("expected 1 route, got %d", len(gw.Routes))
		}
		route := gw.Routes[0]
		if route.Route != "test-subject" || route.Dest != "somewhere" || route.Metric != 3 {
			return fmt.Errorf("route mismatch: %+v", route)
		}
		return nil
	}
}

//go:embed test-data/gatewayconfig/gh159.hcl
var gh159 string

func TestAccGatewayConfig_GH159(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	p := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": p,
		},
		Steps: []resource.TestStep{
			{
				Config: gh159,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_gateway_config.gh159", "enabled", "false"),
					resource.TestCheckResourceAttr("tg_gateway_config.gh159", "node_id", testNodeID),
					resource.TestCheckResourceAttr("tg_gateway_config.gh159", "path.0.id", "somethingsomething-somethingelsesomethingelse-local"),
					resource.TestCheckResourceAttr("tg_gateway_config.gh159", "path.0.host", "10.0.1.18"),
					resource.TestCheckResourceAttr("tg_gateway_config.gh159", "path.0.port", "8443"),
					resource.TestCheckResourceAttr("tg_gateway_config.gh159", "path.0.node", "somethingsomething"),
					checkGatewayConfigGH159APISide(p),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_gateway_config.gh159", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func checkGatewayConfigGH159APISide(p *schema.Provider) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := p.Meta().(*tg.Client)

		var node tg.Node
		if err := client.Get(context.Background(), "/node/"+testNodeID, &node); err != nil {
			return fmt.Errorf("error fetching node: %w", err)
		}
		gw := node.Config.Gateway
		if gw.Enabled {
			return fmt.Errorf("expected gateway disabled")
		}
		if len(gw.Paths) < 1 {
			return fmt.Errorf("expected at least 1 path, got %d", len(gw.Paths))
		}
		path := gw.Paths[0]
		if path.ID != "somethingsomething-somethingelsesomethingelse-local" || path.Host != "10.0.1.18" || path.Port != 8443 || path.Node != "somethingsomething" {
			return fmt.Errorf("path mismatch: %+v", path)
		}
		return nil
	}
}
