package acctests

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/trustgrid/terraform-provider-tg/provider"
)

func TestAccConnector_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider.New("test")(),
		},
		Steps: []resource.TestStep{
			{
				Config: connectorConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_connector.tomcat", "node_id", "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"),
					resource.TestCheckResourceAttr("tg_connector.tomcat", "node", "local"),
					resource.TestCheckResourceAttr("tg_connector.tomcat", "service", "127.0.0.1:8080"),
					resource.TestCheckResourceAttr("tg_connector.tomcat", "port", "8081"),
					resource.TestCheckResourceAttr("tg_connector.tomcat", "protocol", "tcp"),
					resource.TestCheckResourceAttr("tg_connector.tomcat", "description", "tomcat forwarding connector"),
					resource.TestCheckResourceAttr("tg_connector.tomcat", "nic", "eth0"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_connector.tomcat", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func connectorConfig() string {
	return `
resource "tg_connector" "tomcat" {
  node_id     = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
  node        = "local"
  service     = "127.0.0.1:8080"
  port        = 8081
  protocol    = "tcp"
  description = "tomcat forwarding connector"
  nic = "eth0"
}`
}
