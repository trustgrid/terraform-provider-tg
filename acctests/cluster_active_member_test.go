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

// Reuses the same fixture node UID as clusterconfig_test.go. The acctest
// environment is expected to have this node registered.
const activeMemberFixtureNodeID = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"

func TestAccClusterActiveMember_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())
	clusterName := "tf-test-active-member"

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: clusterActiveMemberConfig(clusterName, activeMemberFixtureNodeID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_cluster_active_member.test", "id"),
					resource.TestCheckResourceAttr("tg_cluster_active_member.test", "node_id", activeMemberFixtureNodeID),
					resource.TestCheckResourceAttrSet("tg_cluster_active_member.test", "cluster_fqdn"),
					checkClusterActiveMemberAPISide(provider, activeMemberFixtureNodeID),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_cluster_active_member.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func clusterActiveMemberConfig(clusterName, nodeID string) string {
	return fmt.Sprintf(`
resource "tg_cluster" "test" {
  name = "%s"
}

resource "tg_cluster_member" "test" {
  cluster_fqdn = tg_cluster.test.fqdn
  node_id      = "%s"
}

resource "tg_cluster_active_member" "test" {
  cluster_fqdn = tg_cluster.test.fqdn
  node_id      = "%s"
  depends_on   = [tg_cluster_member.test]
}
	`, clusterName, nodeID, nodeID)
}

// checkClusterActiveMemberAPISide verifies the API-side state: the named
// node should report Config.Cluster.Active=true after the resource creates,
// confirming the underlying `PUT /cluster/{fqdn}/active/{node_id}` succeeded.
func checkClusterActiveMemberAPISide(provider *schema.Provider, nodeID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)

		var node tg.Node
		err := client.Get(context.Background(), fmt.Sprintf("/node/%s", nodeID), &node)
		if err != nil {
			return fmt.Errorf("error getting node %s: %w", nodeID, err)
		}

		if !node.Config.Cluster.Active {
			return fmt.Errorf("expected node %s to be cluster master (Config.Cluster.Active=true), got false", nodeID)
		}

		return nil
	}
}
