package acctests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/trustgrid/terraform-provider-tg/provider"
)

// Cluster data source tests create their own cluster so they don't depend on
// the shared fixture. Node data source tests target testNodeID since there's
// no terraform-managed way to create a fresh node.

func TestAccClusterServicesDataSource(t *testing.T) {
	clusterName := "tf-test-v2-ds-svc-" + acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"tg": provider.New("test")()},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "tg_cluster" "test" {
  name = %q
}

resource "tg_cluster_services_v2_upgrade" "test" {
  cluster_fqdn = tg_cluster.test.fqdn
}

resource "tg_cluster_service" "test" {
  cluster_fqdn = tg_cluster.test.fqdn
  depends_on   = [tg_cluster_services_v2_upgrade.test]
  name         = "ds-test-svc"
  protocol     = "tcp"
  host         = "10.0.0.1"
  port         = 8080
}

data "tg_cluster_services" "all" {
  cluster_fqdn = tg_cluster.test.fqdn
  depends_on   = [tg_cluster_service.test]
}
`, clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.tg_cluster_services.all", "services.#", "1"),
					resource.TestCheckResourceAttr("data.tg_cluster_services.all", "services.0.name", "ds-test-svc"),
				),
			},
		},
	})
}

func TestAccClusterConnectorsDataSource(t *testing.T) {
	clusterName := "tf-test-v2-ds-conn-" + acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"tg": provider.New("test")()},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "tg_cluster" "test" {
  name = %q
}

resource "tg_cluster_connectors_v2_upgrade" "test" {
  cluster_fqdn = tg_cluster.test.fqdn
}

resource "tg_cluster_connector" "test" {
  cluster_fqdn = tg_cluster.test.fqdn
  depends_on   = [tg_cluster_connectors_v2_upgrade.test]
  node         = "local"
  service      = "127.0.0.1:9091"
  port         = 9090
  protocol     = "tcp"
  description  = "ds-test-conn"
  nic          = "any"
}

data "tg_cluster_connectors" "all" {
  cluster_fqdn = tg_cluster.test.fqdn
  depends_on   = [tg_cluster_connector.test]
}
`, clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.tg_cluster_connectors.all", "connectors.#", "1"),
				),
			},
		},
	})
}

// Node data source tests still use the shared testNodeID fixture — no way to
// create a fresh node via the API.

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
