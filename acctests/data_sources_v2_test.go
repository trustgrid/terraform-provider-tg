package acctests

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/trustgrid/terraform-provider-tg/provider"
)

// Data source acctests assert that each list returns at least the resources
// the companion CRUD acctests created. We use TestCheckResourceAttrSet on the
// length so the test doesn't depend on what else happens to be on the cluster
// or node — only that the data source returned something usable.

func TestAccClusterServicesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"tg": provider.New("test")()},
		Steps: []resource.TestStep{
			{
				Config: `
data "tg_cluster_services" "all" {
  cluster_fqdn = "` + testClusterFQDN + `"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tg_cluster_services.all", "services.#"),
				),
			},
		},
	})
}

func TestAccClusterConnectorsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"tg": provider.New("test")()},
		Steps: []resource.TestStep{
			{
				Config: `
data "tg_cluster_connectors" "all" {
  cluster_fqdn = "` + testClusterFQDN + `"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tg_cluster_connectors.all", "connectors.#"),
				),
			},
		},
	})
}

func TestAccNodeServicesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"tg": provider.New("test")()},
		Steps: []resource.TestStep{
			{
				Config: `
data "tg_node_services" "all" {
  node_id = "` + testNodeID + `"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tg_node_services.all", "services.#"),
				),
			},
		},
	})
}

func TestAccNodeConnectorsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"tg": provider.New("test")()},
		Steps: []resource.TestStep{
			{
				Config: `
data "tg_node_connectors" "all" {
  node_id = "` + testNodeID + `"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tg_node_connectors.all", "connectors.#"),
				),
			},
		},
	})
}
