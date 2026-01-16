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

func TestAccClusterConfig_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())
	nodeID := "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: clusterConfigTestConfig(nodeID, true, "10.20.30.40", 7946, "10.20.30.41", 7947),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_node_cluster_config.test", "id", nodeID),
					resource.TestCheckResourceAttr("tg_node_cluster_config.test", "node_id", nodeID),
					resource.TestCheckResourceAttr("tg_node_cluster_config.test", "enabled", "true"),
					resource.TestCheckResourceAttr("tg_node_cluster_config.test", "host", "10.20.30.40"),
					resource.TestCheckResourceAttr("tg_node_cluster_config.test", "port", "7946"),
					resource.TestCheckResourceAttr("tg_node_cluster_config.test", "status_host", "10.20.30.41"),
					resource.TestCheckResourceAttr("tg_node_cluster_config.test", "status_port", "7947"),
					checkClusterConfigAPISide(provider, nodeID, true, "10.20.30.40", 7946, "10.20.30.41", 7947),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_node_cluster_config.test", tfjsonpath.New("id")),
				},
			},
			{
				Config: clusterConfigTestConfig(nodeID, true, "10.20.30.50", 7950, "10.20.30.51", 7951),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_node_cluster_config.test", "id", nodeID),
					resource.TestCheckResourceAttr("tg_node_cluster_config.test", "node_id", nodeID),
					resource.TestCheckResourceAttr("tg_node_cluster_config.test", "enabled", "true"),
					resource.TestCheckResourceAttr("tg_node_cluster_config.test", "host", "10.20.30.50"),
					resource.TestCheckResourceAttr("tg_node_cluster_config.test", "port", "7950"),
					resource.TestCheckResourceAttr("tg_node_cluster_config.test", "status_host", "10.20.30.51"),
					resource.TestCheckResourceAttr("tg_node_cluster_config.test", "status_port", "7951"),
					checkClusterConfigAPISide(provider, nodeID, true, "10.20.30.50", 7950, "10.20.30.51", 7951),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_node_cluster_config.test", tfjsonpath.New("id")),
				},
			},
			{
				Config: clusterConfigTestConfig(nodeID, false, "10.20.30.50", 7950, "10.20.30.51", 7951),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_node_cluster_config.test", "id", nodeID),
					resource.TestCheckResourceAttr("tg_node_cluster_config.test", "node_id", nodeID),
					resource.TestCheckResourceAttr("tg_node_cluster_config.test", "enabled", "false"),
					resource.TestCheckResourceAttr("tg_node_cluster_config.test", "host", "10.20.30.50"),
					resource.TestCheckResourceAttr("tg_node_cluster_config.test", "port", "7950"),
					resource.TestCheckResourceAttr("tg_node_cluster_config.test", "status_host", "10.20.30.51"),
					resource.TestCheckResourceAttr("tg_node_cluster_config.test", "status_port", "7951"),
					checkClusterConfigAPISide(provider, nodeID, false, "10.20.30.50", 7950, "10.20.30.51", 7951),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_node_cluster_config.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func clusterConfigTestConfig(nodeID string, enabled bool, host string, port int, statusHost string, statusPort int) string {
	return fmt.Sprintf(`
resource "tg_node_cluster_config" "test" {
  node_id      = "%s"
  enabled      = %t
  host         = "%s"
  port         = %d
  status_host  = "%s"
  status_port  = %d
}
	`, nodeID, enabled, host, port, statusHost, statusPort)
}

func checkClusterConfigAPISide(provider *schema.Provider, nodeID string, enabled bool, host string, port int, statusHost string, statusPort int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)

		var node tg.Node
		err := client.Get(context.Background(), fmt.Sprintf("/node/%s", nodeID), &node)
		if err != nil {
			return fmt.Errorf("error getting node: %w", err)
		}

		clusterConfig := node.Config.Cluster

		if clusterConfig.Enabled != enabled {
			return fmt.Errorf("expected cluster config enabled to be %t, got %t", enabled, clusterConfig.Enabled)
		}

		if clusterConfig.Host != host {
			return fmt.Errorf("expected cluster config host to be %s, got %s", host, clusterConfig.Host)
		}

		if clusterConfig.Port != port {
			return fmt.Errorf("expected cluster config port to be %d, got %d", port, clusterConfig.Port)
		}

		if clusterConfig.StatusHost != statusHost {
			return fmt.Errorf("expected cluster config status_host to be %s, got %s", statusHost, clusterConfig.StatusHost)
		}

		if clusterConfig.StatusPort != statusPort {
			return fmt.Errorf("expected cluster config status_port to be %d, got %d", statusPort, clusterConfig.StatusPort)
		}

		return nil
	}
}
