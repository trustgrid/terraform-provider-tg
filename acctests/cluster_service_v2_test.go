package acctests

import (
	"context"
	"fmt"
	"os"
	"regexp"
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

// These tests assume the cluster identified by testClusterFQDN has already been
// upgraded to V2 services config. The V2 upgrade endpoint is one-way and
// idempotent server-side, so running these against an unmigrated cluster
// requires a one-time POST /v2/cluster/{fqdn}/config/services/upgrade first.

const v2ClusterTestServiceName = "tf-test-v2-cluster-svc"

func init() {
	resource.AddTestSweepers("tg_cluster_service", &resource.Sweeper{
		Name: "tg_cluster_service",
		F: func(_ string) error {
			cp := tg.ClientParams{
				APIKey:    os.Getenv("TG_API_KEY_ID"),
				APISecret: os.Getenv("TG_API_KEY_SECRET"),
				APIHost:   os.Getenv("TG_API_HOST"),
			}
			client, err := tg.NewClient(context.Background(), cp)
			if err != nil {
				return fmt.Errorf("error creating client: %w", err)
			}

			var cluster tg.Cluster
			if err := client.Get(context.Background(), "/cluster/"+testClusterFQDN, &cluster); err != nil {
				return fmt.Errorf("error fetching cluster %s: %w", testClusterFQDN, err)
			}
			if cluster.Config.Services == nil {
				return nil
			}
			for _, svc := range cluster.Config.Services.Services {
				if svc.Name != v2ClusterTestServiceName {
					continue
				}
				url := fmt.Sprintf("/v2/cluster/%s/config/services/%s", testClusterFQDN, svc.ID)
				if err := client.Delete(context.Background(), url, nil); err != nil {
					return fmt.Errorf("error deleting cluster service %s: %w", svc.ID, err)
				}
			}
			return nil
		},
	})
}

func TestAccClusterServiceV2_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())
	p := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"tg": p},
		Steps: []resource.TestStep{
			{
				Config: clusterServiceV2Config(testClusterFQDN, "", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_cluster_service.test", "service_id"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "cluster_fqdn", testClusterFQDN),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "name", v2ClusterTestServiceName),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "protocol", "tcp"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "host", "10.0.0.1"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "port", "8080"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "enabled", "true"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "source_from_cluster_ip", "false"),
					checkClusterServiceAPISide(p, testClusterFQDN, clusterServiceExpect{
						Name: v2ClusterTestServiceName, Protocol: "tcp", Host: "10.0.0.1", Port: 8080, Enabled: true,
					}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_cluster_service.test", tfjsonpath.New("service_id")),
				},
			},
			{
				Config: clusterServiceV2Config(testClusterFQDN, "ens192", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_cluster_service.test", "source_interface", "ens192"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "source_from_cluster_ip", "true"),
					checkClusterServiceAPISide(p, testClusterFQDN, clusterServiceExpect{
						Name: v2ClusterTestServiceName, Protocol: "tcp", Host: "10.0.0.1", Port: 8080, Enabled: true,
						SourceInterface: "ens192", SourceFromClusterIP: true,
					}),
				),
			},
			{
				Config: clusterServiceV2Config(testClusterFQDN, "ens192", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_cluster_service.test", "source_interface", "ens192"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "source_from_cluster_ip", "false"),
				),
			},
		},
	})
}

func TestAccClusterServiceV2_ValidatorRejectsClusterIPWithoutInterface(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"tg": provider.New("test")()},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "tg_cluster_service" "test" {
  cluster_fqdn           = %q
  name                   = %q
  protocol               = "tcp"
  host                   = "10.0.0.1"
  port                   = 8080
  source_from_cluster_ip = true
}
`, testClusterFQDN, v2ClusterTestServiceName),
				ExpectError: regexp.MustCompile("source_from_cluster_ip = true requires source_interface"),
			},
		},
	})
}

func clusterServiceV2Config(clusterFQDN, sourceInterface string, sourceFromClusterIP bool) string {
	srcLine := ""
	srcFromClusterLine := ""
	if sourceInterface != "" {
		srcLine = fmt.Sprintf(`  source_interface       = %q`, sourceInterface)
		srcFromClusterLine = fmt.Sprintf(`  source_from_cluster_ip = %t`, sourceFromClusterIP)
	}
	return fmt.Sprintf(`
resource "tg_cluster_service" "test" {
  cluster_fqdn = %q
  name         = %q
  protocol     = "tcp"
  host         = "10.0.0.1"
  port         = 8080
  enabled      = true
%s
%s
}
`, clusterFQDN, v2ClusterTestServiceName, srcLine, srcFromClusterLine)
}

type clusterServiceExpect struct {
	Name                string
	Protocol            string
	Host                string
	Port                int
	Enabled             bool
	SourceInterface     string
	SourceFromClusterIP bool
}

func checkClusterServiceAPISide(p *schema.Provider, clusterFQDN string, want clusterServiceExpect) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := p.Meta().(*tg.Client)
		rs, ok := s.RootModule().Resources["tg_cluster_service.test"]
		if !ok {
			return fmt.Errorf("tg_cluster_service.test not found in state")
		}
		serviceID := rs.Primary.ID

		var cluster tg.Cluster
		if err := client.Get(context.Background(), fmt.Sprintf("/cluster/%s", clusterFQDN), &cluster); err != nil {
			return fmt.Errorf("error getting cluster: %w", err)
		}
		if cluster.Config.Services == nil {
			return fmt.Errorf("cluster has no services config")
		}
		var found *tg.Service
		for i := range cluster.Config.Services.Services {
			if cluster.Config.Services.Services[i].ID == serviceID {
				found = &cluster.Config.Services.Services[i]
				break
			}
		}
		if found == nil {
			return fmt.Errorf("service %s not found in cluster config", serviceID)
		}
		if found.Name != want.Name {
			return fmt.Errorf("expected name %q, got %q", want.Name, found.Name)
		}
		if found.Protocol != want.Protocol {
			return fmt.Errorf("expected protocol %q, got %q", want.Protocol, found.Protocol)
		}
		if found.Host != want.Host {
			return fmt.Errorf("expected host %q, got %q", want.Host, found.Host)
		}
		if found.Port != want.Port {
			return fmt.Errorf("expected port %d, got %d", want.Port, found.Port)
		}
		if found.Enabled != want.Enabled {
			return fmt.Errorf("expected enabled %t, got %t", want.Enabled, found.Enabled)
		}
		if found.SourceInterface != want.SourceInterface {
			return fmt.Errorf("expected source_interface %q, got %q", want.SourceInterface, found.SourceInterface)
		}
		if found.SourceFromClusterIP != want.SourceFromClusterIP {
			return fmt.Errorf("expected source_from_cluster_ip %t, got %t", want.SourceFromClusterIP, found.SourceFromClusterIP)
		}
		return nil
	}
}
