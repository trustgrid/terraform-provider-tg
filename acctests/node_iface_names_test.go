package acctests

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/trustgrid/terraform-provider-tg/provider"
)

func TestAccNodeIfaceNames_HappyPath(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider.New("test")(),
		},
		Steps: []resource.TestStep{
			{
				Config: nodeIfaceNamesConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.tg_node_iface_names.test", "id", "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"),
					resource.TestCheckResourceAttrSet("data.tg_node_iface_names.test", "interfaces.0.name"),
					resource.TestCheckResourceAttrSet("data.tg_node_iface_names.test", "interfaces.0.description"),
					resource.TestCheckResourceAttrSet("data.tg_node_iface_names.test", "interfaces.0.os_name"),
				),
			},
		},
	})
}

func nodeIfaceNamesConfig() string {
	return `
data "tg_node_iface_names" "test" {
  node_id = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
}
	`
}
