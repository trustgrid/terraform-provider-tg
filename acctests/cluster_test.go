package acctests

import (
	"context"
	"errors"
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

const testClusterName = "tf-test-cluster"

func init() {
	resource.AddTestSweepers("tg_cluster", &resource.Sweeper{
		Name:         "tg_cluster",
		Dependencies: []string{"tg_cluster_service"},
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

			fqdn := testClusterName + "." + client.Domain
			var cluster tg.Cluster
			if err := client.Get(context.Background(), "/cluster/"+fqdn, &cluster); err != nil {
				var nferr *tg.NotFoundError
				if errors.As(err, &nferr) {
					return nil
				}
				return fmt.Errorf("error checking cluster %s: %w", fqdn, err)
			}
			if err := client.Delete(context.Background(), "/cluster/"+fqdn, nil); err != nil {
				return fmt.Errorf("error deleting cluster %s: %w", fqdn, err)
			}
			return nil
		},
	})
}

func TestAccCluster_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())
	clusterName := testClusterName

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
