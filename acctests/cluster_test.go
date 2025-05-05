package acctests

import (
	"context"
	"fmt"
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

func TestAccCluster_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())
	clusterName := "tf-test-cluster"

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: clusterConfig(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_cluster.test", "id"),
					resource.TestCheckResourceAttr("tg_cluster.test", "name", clusterName),
					resource.TestCheckResourceAttrSet("tg_cluster.test", "fqdn"),
					checkClusterAPISide(provider, clusterName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_cluster.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func clusterConfig(name string) string {
	return fmt.Sprintf(`
resource "tg_cluster" "test" {
  name = "%s"
}
	`, name)
}

func checkClusterAPISide(provider *schema.Provider, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)
		domain := client.Domain

		expectedFQDN := name + "." + domain

		var cluster tg.Cluster
		err := client.Get(context.Background(), "/cluster/"+expectedFQDN, &cluster)
		if err != nil {
			return fmt.Errorf("error getting cluster: %w", err)
		}

		if cluster.Name != name {
			return fmt.Errorf("expected cluster name to be %s, got %s", name, cluster.Name)
		}

		if cluster.FQDN != expectedFQDN {
			return fmt.Errorf("expected cluster FQDN to be %s, got %s", expectedFQDN, cluster.FQDN)
		}

		return nil
	}
}
