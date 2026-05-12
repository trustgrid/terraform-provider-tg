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

const testClusterServiceName = "tf-test-service"

func init() {
	resource.AddTestSweepers("tg_cluster_service", &resource.Sweeper{
		Name: "tg_cluster_service",
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

			var cluster tg.Cluster
			if err := client.Get(context.Background(), "/cluster/"+testClusterFQDN, &cluster); err != nil {
				return fmt.Errorf("error fetching cluster %s: %w", testClusterFQDN, err)
			}
			if cluster.Config.Services == nil {
				return nil
			}
			for id, svc := range cluster.Config.Services.Items {
				if svc.Name != testClusterServiceName {
					continue
				}
				url := fmt.Sprintf("/v2/cluster/%s/config/services/%s", testClusterFQDN, id)
				if err := client.Delete(context.Background(), url, nil); err != nil {
					return fmt.Errorf("error deleting cluster service %s: %w", id, err)
				}
			}
			return nil
		},
	})
}

func TestAccClusterService_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	p := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": p,
		},
		Steps: []resource.TestStep{
			{
				Config: clusterServiceConfig(testClusterFQDN, "", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_cluster_service.test", "service_id"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "cluster_fqdn", testClusterFQDN),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "name", testClusterServiceName),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "protocol", "tcp"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "host", "10.0.0.1"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "port", "8080"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "enabled", "true"),
					resource.TestCheckNoResourceAttr("tg_cluster_service.test", "source_interface"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "source_from_cluster_ip", "false"),
					checkClusterServiceAPISide(p, testClusterFQDN, clusterServiceExpect{
						Name:     testClusterServiceName,
						Protocol: "tcp",
						Host:     "10.0.0.1",
						Port:     8080,
						Enabled:  true,
					}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_cluster_service.test", tfjsonpath.New("service_id")),
				},
			},
			{
				Config: clusterServiceConfig(testClusterFQDN, "ens192", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_cluster_service.test", "service_id"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "source_interface", "ens192"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "source_from_cluster_ip", "true"),
					checkClusterServiceAPISide(p, testClusterFQDN, clusterServiceExpect{
						Name:                testClusterServiceName,
						Protocol:            "tcp",
						Host:                "10.0.0.1",
						Port:                8080,
						Enabled:             true,
						SourceInterface:     "ens192",
						SourceFromClusterIP: true,
					}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_cluster_service.test", tfjsonpath.New("service_id")),
				},
			},
			{
				Config: clusterServiceConfig(testClusterFQDN, "ens192", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_cluster_service.test", "source_interface", "ens192"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "source_from_cluster_ip", "false"),
					checkClusterServiceAPISide(p, testClusterFQDN, clusterServiceExpect{
						Name:            testClusterServiceName,
						Protocol:        "tcp",
						Host:            "10.0.0.1",
						Port:            8080,
						Enabled:         true,
						SourceInterface: "ens192",
					}),
				),
			},
		},
	})
}

func TestAccClusterService_SourceFromClusterIPRequiresInterface(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider.New("test")(),
		},
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
`, testClusterFQDN, testClusterServiceName),
				ExpectError: regexp.MustCompile("source_from_cluster_ip = true requires source_interface"),
			},
		},
	})
}

func clusterServiceConfig(clusterFQDN, sourceInterface string, sourceFromClusterIP bool) string {
	sourceInterfaceLine := ""
	sourceFromClusterIPLine := ""
	if sourceInterface != "" {
		sourceInterfaceLine = fmt.Sprintf(`  source_interface       = %q`, sourceInterface)
		sourceFromClusterIPLine = fmt.Sprintf(`  source_from_cluster_ip = %t`, sourceFromClusterIP)
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
`, clusterFQDN, testClusterServiceName, sourceInterfaceLine, sourceFromClusterIPLine)
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

		svc, found := cluster.Config.Services.Items[serviceID]
		if !found {
			return fmt.Errorf("service %s not found in cluster config (have %d services)", serviceID, len(cluster.Config.Services.Items))
		}

		if svc.Name != want.Name {
			return fmt.Errorf("expected name %q, got %q", want.Name, svc.Name)
		}
		if svc.Protocol != want.Protocol {
			return fmt.Errorf("expected protocol %q, got %q", want.Protocol, svc.Protocol)
		}
		if svc.Host != want.Host {
			return fmt.Errorf("expected host %q, got %q", want.Host, svc.Host)
		}
		if svc.Port != want.Port {
			return fmt.Errorf("expected port %d, got %d", want.Port, svc.Port)
		}
		if svc.Enabled != want.Enabled {
			return fmt.Errorf("expected enabled %t, got %t", want.Enabled, svc.Enabled)
		}
		if svc.SourceInterface != want.SourceInterface {
			return fmt.Errorf("expected source_interface %q, got %q", want.SourceInterface, svc.SourceInterface)
		}
		if svc.SourceFromClusterIP != want.SourceFromClusterIP {
			return fmt.Errorf("expected source_from_cluster_ip %t, got %t", want.SourceFromClusterIP, svc.SourceFromClusterIP)
		}
		return nil
	}
}
