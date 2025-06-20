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

func TestAccCluster_DS_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider.New("test")(),
		},
		Steps: []resource.TestStep{
			{
				Config: clusterDSConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.tg_cluster.test", "name", "test-cluster"),
					resource.TestCheckResourceAttr("data.tg_cluster.test", "fqdn", "test-cluster.terraform.dev.trustgrid.io"),
					resource.TestCheckResourceAttr("data.tg_cluster.test", "health", "offline"),
					resource.TestCheckResourceAttr("data.tg_cluster.test", "members.0.uid", "7ac07330-d2e3-48a4-ad21-1d8d67b6c880"),
					resource.TestCheckResourceAttr("data.tg_cluster.test", "members.0.name", "test-cluster-member"),
					resource.TestCheckResourceAttr("data.tg_cluster.test", "members.0.configured_active", "true"),
					resource.TestCheckResourceAttr("data.tg_cluster.test", "members.0.active", "false"),
					resource.TestCheckResourceAttr("data.tg_cluster.test", "members.0.online", "true"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("data.tg_cluster.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func clusterDSConfig() string {
	return `
data "tg_cluster" "test" {
  fqdn = "test-cluster.terraform.dev.trustgrid.io"
}
	`
}
