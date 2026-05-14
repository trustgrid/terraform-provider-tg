package acctests

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/trustgrid/terraform-provider-tg/provider"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

// TestAccClusterFullLifecycle_V1ToV2 exercises the entire customer migration
// path against a freshly-created cluster in a single test:
//
//   1. V1 path: create legacy tg_service and tg_connector on a fresh V1
//      cluster (using the new provider). Proves backward compatibility.
//   2. Migration: state-remove the V1 resources without API delete, then
//      trigger the V1→V2 upgrade on both services and connectors. Proves the
//      upgrade-trigger resources work and the cluster flips cleanly.
//   3. V2 path: create new V2 services and connectors via tg_cluster_service /
//      tg_cluster_connector, including the V2-only fields. Proves the new
//      resources work on the upgraded cluster.
//   4. Cleanup: TestCase auto-destroy tears down the cluster, which wipes
//      every service and connector inside it — this is the sweeper.
//
// Creating a fresh cluster per test is required because the V1→V2 upgrade is
// irreversible per the API. The shared testClusterFQDN fixture is already V2,
// so the V1 portion of this lifecycle can't be exercised against it.
func TestAccClusterFullLifecycle_V1ToV2(t *testing.T) {
	// Cluster names retain V2 state on the backend even after destroy. A
	// freshly-created cluster with a previously-used name comes back already
	// V2, which breaks the V1 portion of this test. Use a unique suffix.
	clusterName := "tf-test-v2-lc-" + acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)

	p := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"tg": p},
		Steps: []resource.TestStep{
			// Step 1 — V1 path on a freshly-created V1 cluster.
			{
				Config: fullLifecycleV1Config(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_cluster.lc", "fqdn"),
					resource.TestCheckResourceAttrSet("tg_service.v1_svc", "id"),
					resource.TestCheckResourceAttrSet("tg_connector.v1_conn", "id"),
					checkFullLifecycleV1State(p, clusterName),
				),
			},
			// Step 2 — Migration: drop V1 from state without destroying API
			// objects, trigger upgrade.
			{
				Config: fullLifecycleMigrationConfig(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_cluster_services_v2_upgrade.lc", "id"),
					resource.TestCheckResourceAttrSet("tg_cluster_connectors_v2_upgrade.lc", "id"),
					checkClusterIsV2(p, clusterName),
				),
			},
			// Step 3 — V2 path: brand-new V2 resources with V2-only fields.
			{
				Config: fullLifecycleV2Config(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_cluster_service.v2_svc", "service_id"),
					resource.TestCheckResourceAttr("tg_cluster_service.v2_svc", "source_interface", "ens192"),
					resource.TestCheckResourceAttr("tg_cluster_service.v2_svc", "source_from_cluster_ip", "true"),
					resource.TestCheckResourceAttrSet("tg_cluster_connector.v2_conn", "connector_id"),
				),
			},
			// Step 4 — TestCase auto-destroy runs after this; cluster destroy
			// sweeps every service/connector inside it. No explicit step.
		},
	})
}

// Step 1 config: legacy V1 resources on a freshly-created cluster.
func fullLifecycleV1Config(clusterName string) string {
	return fmt.Sprintf(`
resource "tg_cluster" "lc" {
  name = %q
}

resource "tg_service" "v1_svc" {
  cluster_fqdn = tg_cluster.lc.fqdn
  name         = "lifecycle-v1-svc"
  protocol     = "tcp"
  host         = "10.50.50.50"
  port         = 5050
  description  = "V1 service for full lifecycle test"
}

resource "tg_connector" "v1_conn" {
  cluster_fqdn = tg_cluster.lc.fqdn
  node         = "local"
  service      = "127.0.0.1:6060"
  port         = 6060
  protocol     = "tcp"
  description  = "V1 connector for full lifecycle test"
  nic          = "any"
}
`, clusterName)
}

// Step 2 config: state-remove the V1 resources (no API delete) and trigger
// upgrades to V2.
func fullLifecycleMigrationConfig(clusterName string) string {
	return fmt.Sprintf(`
resource "tg_cluster" "lc" {
  name = %q
}

removed {
  from = tg_service.v1_svc
  lifecycle { destroy = false }
}

removed {
  from = tg_connector.v1_conn
  lifecycle { destroy = false }
}

resource "tg_cluster_services_v2_upgrade" "lc" {
  cluster_fqdn = tg_cluster.lc.fqdn
}

resource "tg_cluster_connectors_v2_upgrade" "lc" {
  cluster_fqdn = tg_cluster.lc.fqdn
}
`, clusterName)
}

// Step 3 config: brand-new V2 resources with V2-only fields, plus the upgrade
// resources kept in state so the cluster's V2 state is anchored.
func fullLifecycleV2Config(clusterName string) string {
	return fmt.Sprintf(`
resource "tg_cluster" "lc" {
  name = %q
}

resource "tg_cluster_services_v2_upgrade" "lc" {
  cluster_fqdn = tg_cluster.lc.fqdn
}

resource "tg_cluster_connectors_v2_upgrade" "lc" {
  cluster_fqdn = tg_cluster.lc.fqdn
}

resource "tg_cluster_service" "v2_svc" {
  cluster_fqdn = tg_cluster.lc.fqdn
  depends_on   = [tg_cluster_services_v2_upgrade.lc]
  name         = "lifecycle-v2-svc"
  protocol     = "tcp"
  host         = "10.60.60.60"
  port         = 6060
  enabled      = true

  source_interface       = "ens192"
  source_from_cluster_ip = true
}

resource "tg_cluster_connector" "v2_conn" {
  cluster_fqdn = tg_cluster.lc.fqdn
  depends_on   = [tg_cluster_connectors_v2_upgrade.lc]
  node         = "local"
  service      = "127.0.0.1:7070"
  port         = 7070
  protocol     = "tcp"
  description  = "V2 connector for full lifecycle test"
  enabled      = true
  nic          = "any"
}
`, clusterName)
}

// checkFullLifecycleV1State verifies that step 1 created services/connectors
// on the cluster via the V1 endpoints.
func checkFullLifecycleV1State(p *schema.Provider, clusterName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := p.Meta().(*tg.Client)
		fqdn := clusterName + "." + client.Domain

		var cluster tg.Cluster
		if err := client.Get(context.Background(), "/cluster/"+fqdn, &cluster); err != nil {
			return fmt.Errorf("error getting cluster: %w", err)
		}
		if cluster.Config.Services == nil || len(cluster.Config.Services.Services) == 0 {
			return fmt.Errorf("expected at least one service on cluster after V1 step")
		}
		if cluster.Config.Connectors == nil || len(cluster.Config.Connectors.Connectors) == 0 {
			return fmt.Errorf("expected at least one connector on cluster after V1 step")
		}
		return nil
	}
}

// checkClusterIsV2 asserts that the cluster's services config is in V2 shape
// after the upgrade resource has been applied. The dual-shape decoder
// normalizes both shapes to a slice for consumers, but we can probe the raw
// API behavior: posting a V1-shape body should fail with "Cannot overwrite a
// V2 config", confirming the cluster has flipped.
func checkClusterIsV2(p *schema.Provider, clusterName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := p.Meta().(*tg.Client)
		fqdn := clusterName + "." + client.Domain

		// Try a V1 PUT — on a V2 cluster this returns 422 with "Cannot
		// overwrite a V2 config".
		_, err := client.Put(context.Background(),
			"/cluster/"+fqdn+"/config/services",
			tg.ServicesConfig{Services: []tg.Service{}})
		if err == nil {
			return fmt.Errorf("expected V1 PUT to fail on V2 cluster, but it succeeded — cluster may still be V1")
		}
		return nil
	}
}
