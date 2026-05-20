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

// Second fixture node used for Update + drift tests. This UID is the
// configured-active member of `test-cluster.terraform.dev.trustgrid.io`
// (see cluster_ds_test.go). The Update/drift tests below temporarily move
// it into a test cluster and restore it via t.Cleanup so cluster_ds_test
// keeps passing.
const activeMemberFixtureNodeID2 = "7ac07330-d2e3-48a4-ad21-1d8d67b6c880"
const fixtureNode2OriginalCluster = "test-cluster.terraform.dev.trustgrid.io"

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

// TestAccClusterActiveMember_FailoverStickyAndUpdate validates the two
// behaviors that justify this resource's existence:
//
//  1. Read no-op: an out-of-band failover (master changes via direct API
//     call, simulating a portal click / Lambda failover / node death)
//     is NOT reverted by a subsequent terraform apply with unchanged HCL.
//
//  2. Update: explicitly changing node_id in HCL DOES re-promote, via
//     `PUT /cluster/{fqdn}/active/{node_id}`.
//
// Setup uses both fixture node UIDs as members of a temporary test
// cluster. The second fixture (activeMemberFixtureNodeID2) is the
// configured-active of test-cluster.terraform.dev.trustgrid.io
// (cluster_ds_test.go depends on this) — restored via t.Cleanup.
func TestAccClusterActiveMember_FailoverStickyAndUpdate(t *testing.T) {
	clusterName := "tf-test-active-member-update"

	p := provider.New("test")()

	// Restore activeMemberFixtureNodeID2 to its original cluster after the
	// test so cluster_ds_test continues to pass. The test framework's
	// destroy phase sets cluster=null on the node via tg_cluster_member
	// Delete; this Cleanup re-assigns it to fixtureNode2OriginalCluster.
	t.Cleanup(func() {
		client := p.Meta().(*tg.Client)
		if client == nil {
			return
		}
		_, _ = client.Put(context.Background(),
			fmt.Sprintf("/node/%s", activeMemberFixtureNodeID2),
			map[string]any{"cluster": fixtureNode2OriginalCluster})
		// Best-effort restore of the configured-active designation too.
		_, _ = client.Put(context.Background(),
			fmt.Sprintf("/cluster/%s/active/%s", fixtureNode2OriginalCluster, activeMemberFixtureNodeID2),
			nil)
	})

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": p,
		},
		Steps: []resource.TestStep{
			// Step 1: Create cluster with two members, designate node1 as
			// active. Assert API shows node1 = master.
			{
				Config: twoMemberClusterConfig(clusterName, activeMemberFixtureNodeID, activeMemberFixtureNodeID2, activeMemberFixtureNodeID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_cluster_active_member.test", "node_id", activeMemberFixtureNodeID),
					checkClusterActiveMemberAPISide(p, activeMemberFixtureNodeID),
				),
			},
			// Step 2: Simulate out-of-band failover — directly call
			// PUT /cluster/{fqdn}/active/{node2} via the API client, then
			// re-apply with UNCHANGED HCL (node_id still = node1).
			//
			// Expected: Read no-op keeps state at node1, no plan diff,
			// no PUT call, reality stays on node2.
			{
				PreConfig: func() {
					client := p.Meta().(*tg.Client)
					fqdn := clusterName + "." + client.Domain
					if _, err := client.Put(context.Background(),
						fmt.Sprintf("/cluster/%s/active/%s", fqdn, activeMemberFixtureNodeID2), nil); err != nil {
						t.Fatalf("simulating out-of-band failover: PUT /active/%s failed: %v", activeMemberFixtureNodeID2, err)
					}
				},
				Config: twoMemberClusterConfig(clusterName, activeMemberFixtureNodeID, activeMemberFixtureNodeID2, activeMemberFixtureNodeID),
				// ExpectNonEmptyPlan: false (default) — terraform plan
				// after refresh should be empty, proving Read no-op.
				Check: resource.ComposeTestCheckFunc(
					// HCL still declares node1 — state should still report node1.
					resource.TestCheckResourceAttr("tg_cluster_active_member.test", "node_id", activeMemberFixtureNodeID),
					// API-side reality reflects the out-of-band change (node2 is master).
					checkClusterActiveMemberAPISide(p, activeMemberFixtureNodeID2),
				),
			},
			// Step 3: Update path — explicitly change HCL node_id to node2.
			// Terraform Update should fire PUT /cluster/{fqdn}/active/{node2}.
			// Assert API confirms node2 = master.
			{
				Config: twoMemberClusterConfig(clusterName, activeMemberFixtureNodeID, activeMemberFixtureNodeID2, activeMemberFixtureNodeID2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_cluster_active_member.test", "node_id", activeMemberFixtureNodeID2),
					checkClusterActiveMemberAPISide(p, activeMemberFixtureNodeID2),
				),
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

// twoMemberClusterConfig produces a cluster with two members and one
// active-master designation. `activeNodeID` selects which of the two
// members is declared active in HCL.
func twoMemberClusterConfig(clusterName, node1, node2, activeNodeID string) string {
	return fmt.Sprintf(`
resource "tg_cluster" "test" {
  name = "%s"
}

resource "tg_cluster_member" "node1" {
  cluster_fqdn = tg_cluster.test.fqdn
  node_id      = "%s"
}

resource "tg_cluster_member" "node2" {
  cluster_fqdn = tg_cluster.test.fqdn
  node_id      = "%s"
}

resource "tg_cluster_active_member" "test" {
  cluster_fqdn = tg_cluster.test.fqdn
  node_id      = "%s"
  depends_on   = [tg_cluster_member.node1, tg_cluster_member.node2]
}
	`, clusterName, node1, node2, activeNodeID)
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
