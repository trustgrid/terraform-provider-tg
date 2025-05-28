package acctests

import (
	"testing"

	_ "embed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/trustgrid/terraform-provider-tg/provider"
)

func TestAccGatewayConfig_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider.New("test")(),
		},
		Steps: []resource.TestStep{
			{
				Config: gatewayConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_gateway_config.test", "id", "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"),
					resource.TestCheckResourceAttr("tg_gateway_config.test", "enabled", "true"),
					resource.TestCheckResourceAttr("tg_gateway_config.test", "node_id", "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"),
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
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_gateway_config.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func gatewayConfig() string {
	return `
resource "tg_gateway_config" "test" {
  node_id = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
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
	`
}

//go:embed test-data/gatewayconfig/gh159.hcl
var gh159 string

func TestAccGatewayConfig_GH159(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider.New("test")(),
		},
		Steps: []resource.TestStep{
			{
				Config: gh159,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_gateway_config.gh159", "enabled", "false"),
					resource.TestCheckResourceAttr("tg_gateway_config.gh159", "node_id", "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"),
					resource.TestCheckResourceAttr("tg_gateway_config.gh159", "path.0.id", "somethingsomething-somethingelsesomethingelse-local"),
					resource.TestCheckResourceAttr("tg_gateway_config.gh159", "path.0.host", "10.0.1.18"),
					resource.TestCheckResourceAttr("tg_gateway_config.gh159", "path.0.port", "8443"),
					resource.TestCheckResourceAttr("tg_gateway_config.gh159", "path.0.node", "somethingsomething"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_gateway_config.gh159", tfjsonpath.New("id")),
				},
			},
		},
	})
}
